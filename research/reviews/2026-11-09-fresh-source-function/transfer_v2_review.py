#!/usr/bin/env python3
"""Seal transfer-page judgments before examining any v2 candidate output."""

import argparse
import hashlib
import json
from pathlib import Path


# Judgments follow full-page source-only reading. A claim's factual truth is
# outside this style review; technical uncertainty is not a positive label.
POSITIVES = [
    (1, "number one collection", "unsupported_ranking",
     "An unqualified rank asks readers to accept a comparison without evidence.",
     "Remove the rank; retain that this is a collection of WPF controls, components, and utilities.",
     "The toolkit identity and WPF scope."),
    (1, "next generation Windows applications", "unjustified_intensifier",
     "The generation label adds no capability or version distinction.",
     "Use 'Windows applications'; retain the toolkit's purpose.",
     "The Windows application scope."),
    (2, "Thus, **CoAI.Dev** was born. We hope to create a project that has both a powerful API distribution system and a rich user interface design, becoming the next-generation open-source AIGC project's one-stop commercial solution.",
     "needless_repetition", "The preceding paragraph already states the intended combination and audience; this repeats it with a launch flourish.",
     "Delete this repeated paragraph; retain the preceding concrete rationale and capabilities.",
     "API distribution, user interface, and intended commercial audience in the preceding paragraph."),
    (6, "create the database, create the database and run all migrations:",
     "needless_repetition", "The duplicated instruction interrupts the test setup sequence.",
     "Say 'create the database and run all migrations:' once.",
     "The database creation and migration steps."),
    (7, "next generation grid framework", "unjustified_intensifier",
     "The generation label does not identify an actual successor or measurable capability here.",
     "Use 'grid framework'; retain the responsive and ratio-based details.",
     "The product category and specific responsive-grid distinctions."),
    (7, "Well, a lot of things.", "empty_framing",
     "The answer delays the concrete distinctions in the following sentences.",
     "Delete this sentence; retain the ratio and gutter explanation.",
     "The comparison question and the technical differences that follow."),
    (8, "the best way to build forms with Vue 3", "unsupported_ranking",
     "The banner ranks a linked alternative without stating comparison criteria.",
     "Name FormKit as a Vue 3 option without the rank.",
     "The FormKit link and Vue 3 support."),
    (8, "the easiest way to build forms with Vue", "unsupported_ranking",
     "The absolute ease claim offers no comparison or user-specific evidence.",
     "Introduce Vue Formulate as a way to build forms with Vue.",
     "The product identity, Vue scope, and documentation link."),
    (10, "groundbreaking video foundation model", "unjustified_intensifier",
     "Groundbreaking praises the model without identifying what breaks new ground.",
     "Say 'video foundation model'; retain the model weights and inference-code offer.",
     "The model type and available artifacts."),
    (10, "most advanced open-source human-centric video foundation model", "unsupported_ranking",
     "Most advanced is a global rank not established by the later bounded benchmark.",
     "Remove 'most advanced'; retain the human-centric, open-source model description.",
     "The model scope and benchmark results elsewhere in the page."),
    (10, "The following conclusions can be drawn:  ", "empty_framing",
     "The phrase only introduces the immediately following RTX 4090 bullets.",
     "End the preceding table-summary sentence with a colon and retain both bullets.",
     "The 544p/4-second RTX 4090 test scope and quantitative conclusions."),
    (10, "The following conclusions can be drawn:\n", "empty_framing",
     "The phrase only introduces the immediately following A800 conclusions.",
     "End the preceding table-summary sentence with a colon and retain both A800 conclusions.",
     "The 544p/4-second A800 test scope and quantitative conclusions."),
]

CONTROLS = [
    (1, "It has been downloaded more than 1 million times here, on Codeplex, and on [NuGet](http://www.nuget.org/packages/Extended.Wpf.Toolkit/).", "technical_uncertain", "The numeric download claim requires source verification, not an editorial rank edit."),
    (1, "for non-commercial use", "retain", "The license restriction changes how readers may use the toolkit."),
    (2, "**Rich Model Support**", "retain", "The adjoining provider and format list gives this label concrete scope."),
    (2, "**Zero-cost cross-device conversation sync support for users**", "technical_uncertain", "The cost claim requires verification; the cross-device function is substantive."),
    (2, "the default password `chatnio123456`", "retain", "A default credential is operationally important and must remain visible."),
    (3, "future generations can't access the past", "retain", "The phrase describes the petition's archival concern, not a product-generation claim."),
    (3, "## Next generation community efforts for ActionScript", "retain", "The heading describes successor community work after Flash discontinuation."),
    (3, "__We're not saying Flash and Shockwave player should be preserved as is.__", "retain", "This explicit negation limits the proposal."),
    (4, "The next generation of this technology: [cMeta](https://github.com/cTuningLabs/cmeta)", "retain", "Next generation names a specific successor project."),
    (4, "the next evolution of the [Collective Mind automation framework (MLCommons CM)]", "retain", "The phrase links a successor version to its named predecessor."),
    (5, "very powerful and flexible PHP object mocking\nframework", "optional", "The author's promotional voice accompanies a concrete example and use scope; this is not a clear empty claim."),
    (5, "those object prophecies can't create themselves - there should be a Prophet:", "retain", "The tutorial metaphor introduces a real API object."),
    (5, "The last bit of awesomeness in Prophecy is out-of-the-box spies support.", "optional", "Informal tutorial voice precedes a specific supported feature."),
    (6, "This README is for the unreleased main branch", "retain", "The branch-status warning changes how readers interpret the instructions."),
    (6, "Oban is a powerful and flexible library", "optional", "The nearby job-processing scope and feature details support an authorial description."),
    (6, "The magic happens!", "optional", "This informal tutorial transition precedes the actual execution condition; deleting it is a voice choice."),
    (6, "PostgreSQL 14.0+, MySQL 8.4+, or SQLite3 3.37.0+", "retain", "Version prerequisites must be preserved."),
    (7, "Singularity is [no longer maintained]", "retain", "The project-status warning is essential."),
    (7, "CSS Grid is powerful, and doesn't need a framework like Singularity", "retain", "The statement explains the explicit migration recommendation."),
    (7, "this will *only work with `min-width` media queries!", "retain", "The media-query limit is a technical constraint."),
    (8, "powerful and flexible API", "optional", "This linked claim is inside an HTML feature table and has adjacent feature detail; it is not a clear source-only edit."),
    (8, "all form elements are created with a <a href=\"https://vueformulate.com/guide/#inputs\">single component</a>", "retain", "The component-count property is a concrete capability."),
    (8, "over a dozen languages", "technical_uncertain", "The numerical language claim requires verification and is inside an HTML feature table."),
    (9, "please fill in the following table through `ISSUE`:", "retain", "This instructs contributors how to submit new entries."),
    (9, "🎂🎂 Happy 2nd Birthday to Awesome-LM-SSP! Keep Going! 💪", "retain", "A dated community milestone is intentional project voice."),
    (9, "## Big love to the community — thank you! 🙏", "retain", "An acknowledgment is intentional project voice."),
    (10, "We list the height/width/frame settings we recommend in the following table.", "retain", "The sentence specifies what the table values represent."),
    (10, "**58.3%** compared to HunyuanVideo XDIT (293.3s vs. 464.3s)", "retain", "The measured comparison is substantive evidence with bounded test conditions."),
    (10, "Built upon the open-source framework Diffusers and featuring a non-intrusive parallel implementation approach", "retain", "The named framework and implementation approach are concrete claims."),
    (11, "**This project is end of life**", "retain", "Project support status must remain visible."),
    (11, "you should use the corresponding version of RxJava as the following table:", "retain", "The table provides a conditional version mapping for experimental APIs."),
    (11, "Example usage:", "optional", "The phrase follows a Usage heading but introduces runnable examples; shortening is a voice choice."),
    (12, "All Code Samples target the _Latest GA - 3 Releases_ platform.", "retain", "The release support window is a concrete compatibility promise."),
    (12, "In the following table, you may find all available samples provided in the separated directories as stand-alone projects available for running with the Gradle `runIde` task.", "retain", "The sentence explains table scope and how to run its entries."),
    (12, "These samples show you how features work and help you jumpstart your plugins.", "optional", "The second introduction sentence overlaps the first but adds the plugin-starting purpose."),
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
    if len(source_by_number) != 12 or any(s["status"] != "valid" for s in source_by_number.values()):
        raise ValueError("expected 12 valid pinned sources")
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
            raise ValueError(f"source {number} quote is not unique: {quote[:60]!r}; count={data.count(encoded)}")
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
        judgment["id"] = f"transfer-v2-{index:03d}"
    output = {
        "format": "unswell-349-transfer-v2-source-review-v1",
        "reviewer": "root Codex assistant, source-only; not a human annotator",
        "status": "sealed_before_candidate_output",
        "source_inventory_sha256": hashlib.sha256(inventory_bytes).hexdigest(),
        "scope": "cue_enriched_page_disjoint_transfer_not_prevalence_or_prospective",
        "judgments": judgments,
    }
    args.output.write_text(json.dumps(output, indent=2, ensure_ascii=False) + "\n")
    positives = sum(j["decision"] == "positive" for j in judgments)
    print(f"sealed {positives} positives and {len(judgments) - positives} controls across 12 pages")


if __name__ == "__main__":
    main()
