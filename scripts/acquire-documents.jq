# Acquisition record of one document set; the script supplies every --arg and --argjson.
{
      version: "unswell-corpus-acquisition-v1",
      manifest: {id: $id, seed: $seed,
        weights: {training: 5000, development: 1500, calibration: 1500, final_test: 2000},
        extraction_policy: {contexts: ["comment", "string", "paragraph", "heading", "list-item", "table-cell"], languages: {}, exceptions: []},
        unit_kinds: ["sentence", "paragraph", "fragment"]},
      repository: {name: $name, reference: $reference, commit: $commit, topic: $topic, purpose: $purpose, ecosystem: "text",
        origin: {label: "unknown", scope: "document", evidence: "Dated publication; no unit-level authorship record.", generation_record: ""},
        rights: {license: $license, evidence: "Retained notice LICENSE compiled from the copyright sections of the texts", allowed_uses: ["annotation", "evaluation", "training"]},
        notices: ["LICENSE"],
        snapshot: {date: $date, confidence: "corroborated", evidence: $evidence, cohort: $cohort}},
      selection: {max_sources: 400, shard_sources: 1, shard_bytes: 1048576, max_source_bytes: 1048576, document_roots: [], document_role: "specification",
        source_extensions: [],
        excluded_segments: $segments, excluded_basenames: $basenames, generated_markers: $markers, translation_hints: $hints}
    }
