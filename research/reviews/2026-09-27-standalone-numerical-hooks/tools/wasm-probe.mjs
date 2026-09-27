import assert from 'node:assert/strict';
import { readFile, writeFile } from 'node:fs/promises';
import { createHash } from 'node:crypto';
import { pathToFileURL } from 'node:url';
const { boot } = await import(pathToFileURL(process.argv[2]).href);
const page = await readFile(process.argv[3], 'utf8');
const runtime = await boot();
const id = 'filler.unnamed-numerical-choice';
const original = 'Four answers, and the right one depends on whether you control the writes.';
const revision = page.replace(original, 'Choose a consistency mode based on whether you control source writes.');
const rows = [];
for (const profile of ['technical', 'strict']) {
  const report = await runtime.analyze(page, profile);
  assert.equal(report.incomplete, false);
  assert.equal(report.gate.passed, true);
  const findings = report.findings.filter(f => f.ruleId === id);
  assert.equal(findings.length, 1);
  const f = findings[0];
  assert.equal(f.start, 477);
  assert.equal(f.end, 550);
  assert.equal(Buffer.from(page).subarray(f.start, f.end).toString(), original.slice(0, -1));
  assert.equal(f.severity, 'warning');
  assert.ok(f.suggestion.includes('Preserve the stated condition'));
  for (const text of [revision,
    'Two schemes: env and file.',
    'Four migration paths are available: validate, baseline, checkpoint, and replay.',
    'Four answers are required by the protocol.',
    '## Four options for configuring a provider',
    'Four options, and the right one depends on the mode: local, hosted, gateway, or offline.',
    'Four answers, and the right one depends on the mode; local or hosted.']) {
    const control = await runtime.analyze(text, profile);
    assert.equal(control.incomplete, false);
    assert.equal(control.findings.filter(f => f.ruleId === id).length, 0);
  }
  rows.push({profile, matched: 1, start: f.start, end: f.end, gate: report.gate, controls: 7});
}
assert.deepEqual(runtime.panics, []);
assert.equal(runtime.info.commit, runtime.manifest.commit);
const result = {candidate: runtime.info.commit, source_sha256: createHash('sha256').update(page).digest('hex'), rows};
await writeFile(process.argv[4], JSON.stringify(result, null, 2) + '\n');
console.log(JSON.stringify(result));
process.exit(0);
