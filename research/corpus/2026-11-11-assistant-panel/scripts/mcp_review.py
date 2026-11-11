#!/usr/bin/env python3
"""Run review in context through the MCP server with one calling assistant.

usage: mcp_review.py <caller> <workspace> <server binary> <out dir> <docs.jsonl>... [--pages N] [--workers N]

For every sampled page the client asks the server for the rubric and the
items (unswell_review_request), gives both to the calling assistant unchanged,
and submits the assistant's flags (unswell_review_submit). The server binds
each flag to source bytes or rejects it. Nothing is repaired by hand.
"""
import argparse, concurrent.futures, hashlib, json, os, subprocess, sys, tempfile, threading, time, urllib.request
from pathlib import Path

import annotate as base

SCHEMA = {"type": "object", "additionalProperties": False, "required": ["flags"],
          "properties": {"flags": {"type": "array", "items": {
              "type": "object", "additionalProperties": False, "required": ["item", "categories", "quote", "reason"],
              "properties": {"item": {"type": "string"}, "categories": {"type": "array", "items": {"type": "string"}},
                             "quote": {"type": "string"}, "reason": {"type": "string"}}}}}}
ANSWER = ('\nAnswer with JSON only: {"flags": [{"item": "<id>", "categories": ["<category>"], '
          '"quote": "<text copied from the item>", "reason": "<one sentence>"}]}. '
          'Return {"flags": []} when nothing needs revision.\n')


class Server:
    """One MCP server process, spoken to over newline-delimited JSON-RPC."""

    def __init__(self, binary):
        self.process = subprocess.Popen([binary, "--timeout", "5m"], stdin=subprocess.PIPE, stdout=subprocess.PIPE,
                                        stderr=subprocess.DEVNULL)
        self.next_id = 0
        self.request("initialize", {"protocolVersion": "2025-06-18", "capabilities": {},
                                    "clientInfo": {"name": "corpus-review-client", "version": "1"}})
        self.notify("notifications/initialized", {})

    def send(self, message):
        self.process.stdin.write((json.dumps(message) + "\n").encode())
        self.process.stdin.flush()

    def notify(self, method, params):
        self.send({"jsonrpc": "2.0", "method": method, "params": params})

    def request(self, method, params):
        self.next_id += 1
        self.send({"jsonrpc": "2.0", "id": self.next_id, "method": method, "params": params})
        while True:
            line = self.process.stdout.readline()
            if not line:
                raise ValueError("server closed the connection")
            message = json.loads(line)
            if message.get("id") == self.next_id:
                if "error" in message:
                    raise ValueError("server error: " + json.dumps(message["error"])[:300])
                return message["result"]

    def tool(self, name, arguments):
        result = self.request("tools/call", {"name": name, "arguments": arguments})
        if result.get("isError"):
            raise ValueError("tool error: " + json.dumps(result.get("content"))[:300])
        if "structuredContent" in result:
            return result["structuredContent"]
        return json.loads(result["content"][0]["text"])

    def close(self):
        self.process.stdin.close()
        self.process.wait(timeout=10)


def ask(spec, prompt, scratch):
    if spec["kind"] == "codex":
        schema = Path(scratch) / "schema.json"
        schema.write_text(json.dumps(SCHEMA))
        answer = Path(scratch) / "answer.json"
        args = ["codex", "exec", *base.CODEX_OPTIONS, "--model", spec["model"], "--cd", scratch,
                "--output-schema", str(schema), "--output-last-message", str(answer),
                "-c", f'model_reasoning_effort="{spec["effort"]}"']
        for setting in base.CODEX_SETTINGS:
            args += ["-c", setting]
        base.run(args + ["-"], prompt.encode(), scratch)
        return answer.read_text()
    if spec["kind"] == "claude":
        args = ["claude", "-p", "--model", spec["model"], "--tools", "", "--strict-mcp-config",
                "--setting-sources", "", "--no-session-persistence", "--output-format", "json",
                "--json-schema", json.dumps(SCHEMA),
                "--system-prompt", "You answer with the requested JSON object and nothing else."]
        out = json.loads(base.run(args, prompt.encode(), scratch))
        if out.get("is_error"):
            raise ValueError("claude error: " + str(out.get("result"))[:200])
        structured = out.get("structured_output")
        return json.dumps(structured) if structured is not None else out["result"]
    body = {"model": os.environ["ROUTER_MODEL"], "temperature": 0, "max_tokens": 24000, "stream": False,
            "messages": [{"role": "user", "content": prompt}]}
    request = urllib.request.Request(os.environ["ROUTER_URL"], data=json.dumps(body).encode(),
                                     headers={"Content-Type": "application/json",
                                              "Authorization": "Bearer " + os.environ["ROUTER_KEY"]})
    with urllib.request.urlopen(request, timeout=base.TIMEOUT) as response:
        choice = json.load(response)["choices"][0]
    if not choice["message"].get("content"):
        raise ValueError("empty answer, finish reason " + str(choice.get("finish_reason")))
    return choice["message"]["content"]


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("caller", choices=sorted(base.ANNOTATORS))
    parser.add_argument("workspace")
    parser.add_argument("server")
    parser.add_argument("out_dir")
    parser.add_argument("docs", nargs="+")
    parser.add_argument("--pages", type=int, default=200)
    parser.add_argument("--workers", type=int, default=4)
    options = parser.parse_args()
    W = Path(options.workspace)
    spec = base.ANNOTATORS[options.caller]
    documents = []
    for name in options.docs:
        folder = Path(name).stem.removeprefix("docs-")
        for line in open(W / name):
            doc = json.loads(line)
            documents.append((hashlib.sha256(doc["doc_id"].encode()).hexdigest(), folder, doc))
    documents = sorted(documents, key=lambda entry: entry[0])[:options.pages]
    os.makedirs(options.out_dir, exist_ok=True)
    local = threading.local()

    def review(entry):
        _, folder, doc = entry
        target = Path(options.out_dir) / f"{folder}__{doc['doc_id']}.json"
        if target.exists():
            return "cached"
        if not hasattr(local, "server"):
            local.server = Server(options.server)
        source = {"name": doc["doc_id"] + ".md", "format": "markdown", "text": doc["text"]}
        last = None
        for attempt in (1, 2, 3):
            started = time.time()
            try:
                asked = local.server.tool("unswell_review_request", {"source": source})
                lines = [asked["instructions"], "\n===== DOCUMENT =====\n", doc["text"], "\n===== ITEMS =====\n"]
                for item in asked["items"]:
                    lines.append(f"{item['id']} (line {item['line']}): {item['text']}\n")
                prompt = "".join(lines) + ANSWER
                with tempfile.TemporaryDirectory() as scratch:
                    answer = base.parse(ask(spec, prompt, scratch))
                flags = [{"item": str(f.get("item", "")), "categories": [str(c) for c in f.get("categories", [])],
                          "quote": str(f.get("quote", "")), "reason": str(f.get("reason", ""))}
                         for f in answer["flags"]][:256]
                bound = local.server.tool("unswell_review_submit", {"source": source, "flags": flags}) if flags else \
                    {"accepted": [], "rejected": []}
                record = {"path": f"{folder}/{doc['doc_id']}.md", "caller": options.caller, "model": spec["model"],
                          "protocol": asked["protocol"], "items": len(asked["items"]), "flags": len(flags),
                          "accepted": bound.get("accepted") or [], "rejected": bound.get("rejected") or [],
                          "prompt_sha256": hashlib.sha256(prompt.encode()).hexdigest(),
                          "seconds": round(time.time() - started, 1)}
                target.with_suffix(".tmp").write_text(json.dumps(record, ensure_ascii=False, indent=1))
                target.with_suffix(".tmp").rename(target)
                return f"ok {len(flags)} flags, {len(record['accepted'])} bound {record['seconds']}s"
            except Exception as error:  # noqa: BLE001
                last = f"{type(error).__name__}: {str(error)[:200]}"
                if "server" in str(error):
                    local.server = Server(options.server)
        return "FAILED " + str(last)

    with concurrent.futures.ThreadPoolExecutor(options.workers) as pool:
        for entry, outcome in zip(documents, pool.map(review, documents)):
            print(f"{entry[1]}/{entry[2]['doc_id']} {outcome}", flush=True)


if __name__ == "__main__":
    main()
