#!/usr/bin/env python3
"""Seal source-only editorial judgments with original UTF-8 byte spans."""

import argparse
import hashlib
import json
from pathlib import Path


# Each judgment was made after reading the complete pinned README and before
# running a candidate detector. Short spans identify the actual edit target;
# the explanation records what must survive any edit.
POSITIVES = [
    (1, "4)The feature string may be changed along with the evolution of hardening services.",
     "needless_repetition", "Tip 4 repeats the opening notice without adding a step.",
     "Delete tip 4; retain the opening notice and its failure consequence.",
     "The feature string can change and a mismatch prevents unpacking."),
    (1, "DexHunter has its own limitation.",
     "empty_framing", "The sentence announces a limitation that the next sentence specifies.",
     "Delete this sentence and keep the explanation that follows.",
     "Hardening changes may make DexHunter ineffective."),
    (3, "next-generation AI doctor model",
     "unjustified_intensifier", "Next-generation adds promotion without an observable capability.",
     "AI doctor model", "The project identity and its LLaMA basis."),
    (3, "The ChatDoctor is an advanced language model that is specifically designed for medical applications.",
     "vague_claim", "Advanced is unevidenced and the medical purpose is already stated.",
     "ChatDoctor is a language model designed for medical applications.",
     "The medical application and following training details."),
    (5, "The installation is quite trivial, for details visit [official site](https://help.ubuntu.com/community/Installation/MinimalCD).",
     "unjustified_intensifier", "Trivial judges the reader's task without helping them install it.",
     "For installation details, visit the [official site](https://help.ubuntu.com/community/Installation/MinimalCD).",
     "The link to installation instructions."),
    (5, "This will install curl and git to your system.",
     "needless_repetition", "The immediately preceding command already names both packages.",
     "Delete this sentence.", "The `apt-get install git curl` command."),
    (5, "[Sphinx](http://sphinxsearch.com/) is powerful and flexible search engine used in Openbay.",
     "vague_claim", "Two broad compliments do not explain the search engine's role.",
     "[Sphinx](http://sphinxsearch.com/) is the search engine used in Openbay.",
     "The Sphinx name, link, and use in Openbay."),
    (6, "best-of-class",
     "vague_claim", "The ranking claim has no comparison or evidence here.",
     "Delete best-of-class from the Tendermint description.",
     "Tendermint is the BFT proof-of-stake consensus engine used."),
    (6, "While both of those are extremely powerful and flexible\nand contain advanced features, they have a steep learning\ncurve for novice users.",
     "unjustified_intensifier", "The stacked praise obscures the concrete learning-curve tradeoff.",
     "Both SDK releases have advanced features, but a steep learning curve for novice users.",
     "The comparison to Cosmos SDK releases and their learning curve."),
    (8, "It also has many outstanding advantages such as cross-platform, high performance, model compression, and code tailoring.",
     "vague_claim", "Outstanding praises the listed capabilities without evidence.",
     "Its capabilities include cross-platform inference, performance optimization, model compression, and code tailoring.",
     "The four listed capabilities."),
    (8, "The support for each demo is shown in the following table.",
     "empty_framing", "The sentence only announces the immediately adjacent table.",
     "Delete this sentence; keep the next sentence explaining clickable checkmarks.",
     "The table and its clickable entry-code links."),
    (8, "It is very simple to use TNN.",
     "vague_claim", "The ease claim precedes actual numbered steps and adds no instruction.",
     "Delete this sentence.", "The three deployment steps."),
    (8, "We provide a wealth of tools to help you complete this step, whether you are using Tensorflow, Pytorch, or Caffe, you can easily complete the conversion.",
     "unjustified_intensifier", "Wealth and easily overpromise where supported frameworks suffice.",
     "TNN provides conversion tools for TensorFlow, PyTorch, and Caffe.",
     "The supported input frameworks and conversion tools."),
    (8, "At present, TNN has been launched in various major businesses, and its following characteristics have been widely praised.",
     "vague_claim", "Unspecified businesses and praise supply no evidence for the following technical list.",
     "Introduce the following characteristics directly; retain the named deployments in the introduction.",
     "The technical characteristics and named applications earlier in the page."),
    (8, "Everyone is welcome to participate to build the best inference framework in the industry.",
     "unjustified_intensifier", "Best in the industry is an unsupported ranking attached to a contribution request.",
     "Contributions are welcome.", "The invitation to contribute."),
    (9, "next generation gaming digital economy platform",
     "unjustified_intensifier", "Next generation promotes without specifying a capability.",
     "gaming digital economy platform", "The platform identity and following goals."),
    (9, "rich api connection support",
     "vague_claim", "Rich adds no information beyond the listed SDKs.",
     "API support", "The JavaScript, iOS, Android, and Python SDK list."),
]

CONTROLS = [
    (1, 'It is worth noting that some "annotation_off" or "debug_info_off" fields may be invalid in the result.',
     "optional", "The lead-in may be shortened, but the specific offset warning is useful."),
    (1, "This string is very important. If it is incorrect, the unpacking process will fail.",
     "retain", "The historical comment gives a changed string and the failure consequence."),
    (2, "No configuration or central generator server\nis required so it can be used directly in server's code.",
     "retain", "This is a concrete deployment distinction from Snowflake."),
    (2, "It is worth noting that most other UUID-like implementations are also not cryptographically secure.",
     "optional", "The opening may be shortened, but the comparison adds a useful security caveat."),
    (2, "Unicity guaranteed for 16,777,216 (24 bits) unique ids per second and per host/process",
     "technical_uncertain", "The numerical guarantee requires technical verification, not a style edit."),
    (2, "Note: UUIDv1 requires a global lock, hence the performance degradation as we add more CPUs.",
     "retain", "The note explains the benchmark trend with a specific mechanism."),
    (3, "please do not apply it to real clinical scenarios.",
     "retain", "This safety warning must survive the lead-in edit."),
    (3, "It is worth noting that our model has not yet achieved 100% accurate output, please do not apply it to real clinical scenarios.",
     "optional", "Direct wording may be preferable, but a clinical-use warning is not a clear empty wrapper."),
    (3, "It has been trained on a large corpus of medical literature and has a deep understanding of medical terminology, procedures, and diagnoses.",
     "technical_uncertain", "Training and claimed capability require external evidence before rewriting."),
    (3, "One of the key features of the ChatDoctor model is its ability to learn and adapt over time.",
     "technical_uncertain", "Whether deployment updates the model requires technical evidence."),
    (3, "ChatDoctor accurately mentions sinusitis, which is also present in the real conversation in icliniq.",
     "retain", "The sentence points to the displayed example's specific overlap."),
    (4, "Under the hood Auto Layout is a powerful and flexible way of organising and laying out your views.",
     "optional", "The praise frames a contrast with verbose constraint creation; removing it is a voice choice."),
    (4, "Masonry supports iOS and Mac OS X.",
     "retain", "Platform support is a concrete product scope."),
    (4, "Masonry however will automagically add constraints to the appropriate view.",
     "retain", "The informal voice accompanies a specific behavior demonstrated nearby."),
    (5, "Openbay is not supported",
     "retain", "The project status is the page's central warning."),
    (5, "Due to bower assets it can take long time.",
     "retain", "The caution prepares readers for a slow dependency step."),
    (5, "This manual describes how to install php-fpm. If you want to use FastCGI refer to PHP documentation.",
     "retain", "The text establishes which server setup this guide covers."),
    (6, "maintaining multiple merkle stores in one app",
     "retain", "A specific limit supports the advice to switch frameworks."),
    (6, "Note that this app relies on a separate tendermint process\nto drive it.",
     "retain", "The operational prerequisite must remain visible."),
    (6, "The end goal is to make blockchain development almost as\nproductive as web development",
     "optional", "An explicitly aspirational goal explains the author's design intent."),
    (7, "Operators using numeric values will typically also operate on letters and convert them into values as per the following table.",
     "retain", "The sentence states what the adjacent base36 table maps."),
    (7, 'Note that `msb` may also be identified as "bank" and `lsb` as "sub" in some applications (like Ableton Live).',
     "retain", "The note gives domain-specific aliases needed to use the command."),
    (7, "This application **is not a synthesizer, but a livecoding environment**",
     "retain", "The negation prevents a wrong expectation about the product."),
    (8, "You can click the ✅ and find the entrance code for each demo.",
     "retain", "The sentence explains that table cells link to example code."),
    (8, "TNN supports INT8/FP16 mode, reduces model size & memory consumption, and utilizes specific hardware low-precision instructions to accelerate calculations.",
     "retain", "The numerical modes and hardware mechanism are substantive."),
    (8, "The size of the mobile dynamic library is only around 400KB",
     "technical_uncertain", "The size claim needs measurement before any factual change."),
    (9, "Consider the following table as minimal requirements before running a node:",
     "retain", "Minimal establishes how to interpret the hardware table."),
    (9, "a full node will need more than 16GB of RAM to operate",
     "retain", "The dated capacity warning is operationally important."),
    (9, "If you send private keys over this connection, rpc-endpoint should be bound to localhost for security.",
     "retain", "Removing this warning could expose credentials."),
    (9, "The project aims to provide game developers with an easy-to-use, comprehensive blockchain gaming infrastructure.",
     "optional", "This is a stated product goal, not a proven capability claim."),
]


def main() -> None:
    parser = argparse.ArgumentParser()
    parser.add_argument("--sources-dir", type=Path, required=True)
    parser.add_argument("--inventory", type=Path, required=True)
    parser.add_argument("--output", type=Path, required=True)
    args = parser.parse_args()
    if args.output.exists():
        raise ValueError("sealed review already exists")
    inventory_bytes = args.inventory.read_bytes()
    inventory = json.loads(inventory_bytes)
    source_by_number = {source["number"]: source for source in inventory["sources"]}
    if len(source_by_number) != 9 or any(s["status"] != "valid" for s in source_by_number.values()):
        raise ValueError("expected nine valid pinned sources")
    pages = {}
    for number, source in source_by_number.items():
        data = (args.sources_dir / source["source_file"]).read_bytes()
        if hashlib.sha256(data).hexdigest() != source["source_sha256"]:
            raise ValueError(f"source {number} hash changed")
        pages[number] = data
    judgments = []

    def append(number: int, quote: str, decision: str, fields: dict) -> None:
        data = pages[number]
        encoded = quote.encode("utf-8")
        if data.count(encoded) != 1:
            raise ValueError(f"source {number} quote is not unique: {quote[:60]!r}")
        start = data.index(encoded)
        judgments.append({"page": number, "source_sha256": source_by_number[number]["source_sha256"],
                          "start_byte": start, "end_byte": start + len(encoded),
                          "quote": quote, "decision": decision, **fields})

    for number, quote, category, burden, edit, retained in POSITIVES:
        append(number, quote, "positive", {"category": category,
               "reader_burden": burden, "meaning_preserving_edit": edit,
               "content_to_retain": retained, "edit_safe": True})
    for number, quote, decision, reason in CONTROLS:
        append(number, quote, decision, {"reason": reason})
    judgments.sort(key=lambda item: (item["page"], item["start_byte"], item["end_byte"]))
    for index, judgment in enumerate(judgments, 1):
        judgment["id"] = f"fresh-{index:03d}"
    output = {
        "format": "unswell-349-fresh-source-function-review-v2",
        "reviewer": "root Codex assistant, source-only; not a human annotator",
        "status": "sealed_before_candidate_output",
        "source_inventory_sha256": hashlib.sha256(inventory_bytes).hexdigest(),
        "scope": "cue_enriched_not_prevalence_or_prospective",
        "judgments": judgments,
    }
    args.output.write_text(json.dumps(output, indent=2, ensure_ascii=False) + "\n")
    positives = sum(j["decision"] == "positive" for j in judgments)
    print(f"sealed {positives} positives and {len(judgments) - positives} controls across 9 pages")


if __name__ == "__main__":
    main()
