#!/usr/bin/env python3
"""Run `unswell compare` on each page and its revision, and keep the verdict.

usage: compare.py <unswell binary> <judge> <cache dir> <out jsonl> <revisions jsonl>... [--workers N] [--only set,...]
                         [--limit N]

A revisions file holds rows with "set", "id", "original", and "revised". The judge is "router" for the model that
ROUTER_URL, ROUTER_MODEL, and ROUTER_KEY name in the environment, or a kind of ask_model.py, such as astra-low, which
is then started as a local program for each question. Answers of a local program are cached in the cache directory.

A row of the output holds the set and id of the page, whether the revision was accepted, the reasons, the verdict of
the blind comparison with both answers, the statements with their outcomes, what was lost, and the em dashes of
both texts. A comparison that did not complete has "error" and no verdict.
"""
import concurrent.futures, json, os, subprocess, sys, tempfile
from pathlib import Path

binary, judge, cache, out = sys.argv[1], sys.argv[2], sys.argv[3], Path(sys.argv[4])
option = lambda name, default: sys.argv[sys.argv.index(name) + 1] if name in sys.argv else default
workers, limit = int(option("--workers", "4")), int(option("--limit", "100000"))
only = set(option("--only", "").split(",")) - {""}
WORKSPACE = str(Path(__file__).resolve().parents[2])
# The prompts that this run has asked; a prompt that comes again is a repeated question and gets a new answer.
RUN = tempfile.mkdtemp(prefix="compare-run-")


def model_arguments():
    if judge == "router":
        return ["--endpoint", os.environ["ROUTER_URL"], "--model", os.environ["ROUTER_MODEL"], "--api-key-env", "ROUTER_KEY"]
    return ["--command", sys.executable, "--command-arg", str(Path(__file__).with_name("ask_repeat.py")), "--command-arg", WORKSPACE + "/ask_model.py",
            "--command-arg", judge, "--command-arg", cache, "--command-arg", RUN]


def compare(row):
    with tempfile.TemporaryDirectory() as scratch:
        page, revised, report = Path(scratch) / "page.md", Path(scratch) / "revised.md", Path(scratch) / "report.json"
        page.write_text(row["original"])
        revised.write_text(row["revised"])
        done = subprocess.run([binary, "compare", str(page), str(revised), "--quiet", "--report", "json:" + str(report), "--request-timeout", "10m",
                               "--timeout", "90m", "--attempts", "3", *model_arguments()], cwd=scratch, capture_output=True)
        result = json.loads(report.read_text()) if report.exists() else {"error": done.stderr.decode()[-400:]}
    kept = {"set": row["set"], "id": row["id"], "exit": done.returncode}
    if result.get("error") or "verdict" not in result:
        kept["error"] = result.get("error") or done.stderr.decode()[-400:]
        return kept
    # A build with two results says apart whether nothing is lost; an earlier build has only the reasons.
    if "kept" in result["verdict"]:
        kept.update(require=result["verdict"]["require"], nothing_lost=result["verdict"]["kept"]["passed"],
                    lost_reasons=result["verdict"]["kept"]["reasons"])
    kept.update(accepted=result["verdict"]["accepted"], reasons=result["verdict"]["reasons"], sound=result["sound"]["verdict"],
                named=[result["sound"]["page_first"]["named"], result["sound"]["revision_first"]["named"]],
                reason=[result["sound"]["page_first"]["reason"], result["sound"]["revision_first"]["reason"]],
                facts=[[f["outcome"], f["text"], f.get("quote", "")] for f in result["facts"]["statements"]], dropped=len(result["facts"].get("dropped") or []),
                lost=result["lost"],
                dashes=[result["em_dashes"]["page"], result["em_dashes"]["revision"]])
    # A build that compares the sentences of the page with the revision reports the suspects and what the judge said of each.
    if "sentences" in result:
        kept.update(compared=result["sentences"]["compared"],
                    sentences=[[s["reason"], s.get("outcome", ""), s["text"], s.get("quote", ""), s.get("cut", ""), bool(s.get("padding")),
                                s.get("revised", "") != ""] for s in result["sentences"]["suspects"]])
    if "damage" in row:
        kept["damage"] = row["damage"]
    return kept


done = {(r["set"], r["id"]) for r in map(json.loads, out.read_text().splitlines()) if "error" not in r} if out.exists() else set()
jobs = []
for path in [a for a in sys.argv[5:] if not a.startswith("--") and a not in (option("--workers", ""), option("--only", ""), option("--limit", ""))]:
    for row in map(json.loads, open(path)):
        if "revised" in row and (not only or row["set"] in only) and (row["set"], row["id"]) not in done:
            jobs.append(row)
jobs = jobs[:limit]
with out.open("a") as handle, concurrent.futures.ThreadPoolExecutor(workers) as pool:
    for row in pool.map(compare, jobs):
        handle.write(json.dumps(row, ensure_ascii=False) + "\n")
        handle.flush()
print("compared", len(jobs))
