#!/usr/bin/env node
// Writes the two files that describe the assembled Pages root: versions.json,
// which the version picker reads, and index.html, the apex redirect.
//
// The assembly holds one directory per documentation version. This script reads
// the directory names rather than a list kept by hand, so a version that failed
// to build is a version the picker does not offer and the redirect cannot land
// on.
import { existsSync, mkdtempSync, mkdirSync, readdirSync, readFileSync, rmSync, writeFileSync } from 'node:fs';
import { tmpdir } from 'node:os';
import { join } from 'node:path';

// The site root. Empty: the documentation is served at the apex of its own
// domain, so there is nothing between the root and a version folder.
export const PAGES_PREFIX = '';

const EDGE = 'edge';
// Released tags carry a prerelease suffix during the alpha, so the pattern has
// to accept one. Build metadata is deliberately not accepted: it does not
// participate in precedence and would make two folders sort equal.
const VERSION_RE = /^v(\d+)\.(\d+)(?:\.(\d+))?(?:-([0-9A-Za-z][0-9A-Za-z.-]*))?$/;

export function isVersionFolder(name) {
	return name === EDGE || VERSION_RE.test(name);
}

export function parseSemver(name) {
	const match = VERSION_RE.exec(name);
	if (!match) return null;
	return {
		numbers: [Number(match[1]), Number(match[2]), match[3] === undefined ? 0 : Number(match[3])],
		prerelease: match[4] === undefined ? [] : match[4].split('.'),
	};
}

// comparePrerelease implements semver precedence for the dotted suffix: a
// version with a prerelease ranks below the same version without one, numeric
// identifiers compare numerically, and a longer list wins a tie.
function comparePrerelease(a, b) {
	if (a.length === 0 && b.length === 0) return 0;
	if (a.length === 0) return 1;
	if (b.length === 0) return -1;
	for (let i = 0; i < Math.min(a.length, b.length); i += 1) {
		const left = a[i];
		const right = b[i];
		const leftNumeric = /^\d+$/.test(left);
		const rightNumeric = /^\d+$/.test(right);
		if (leftNumeric && rightNumeric) {
			if (Number(left) !== Number(right)) return Number(left) - Number(right);
			continue;
		}
		if (leftNumeric !== rightNumeric) return leftNumeric ? -1 : 1;
		if (left !== right) return left < right ? -1 : 1;
	}
	return a.length - b.length;
}

export function compareSemver(a, b) {
	for (let i = 0; i < 3; i += 1) {
		if (a.numbers[i] !== b.numbers[i]) return a.numbers[i] - b.numbers[i];
	}
	return comparePrerelease(a.prerelease, b.prerelease);
}

// computeDefault names the version the apex redirect serves.
//
// The newest released tag, not edge. A reader who types the bare domain has
// installed something, or is about to, and what they can install is a release.
// Edge stays one selection away in the picker and keeps its own stable address.
// Edge is the answer only when the assembly holds no tag at all, which is what
// a first deploy from the default branch looks like.
export function computeDefault(slugs) {
	let best = null;
	for (const slug of slugs) {
		const semver = parseSemver(slug);
		if (semver && (best === null || compareSemver(semver, best.semver) > 0)) {
			best = { slug, semver };
		}
	}
	return best ? best.slug : EDGE;
}

export function buildIndex(slugs) {
	const tags = slugs
		.filter((slug) => parseSemver(slug))
		.sort((a, b) => compareSemver(parseSemver(b), parseSemver(a)));
	const ordered = [];
	if (slugs.includes(EDGE)) ordered.push(EDGE);
	ordered.push(...tags);
	return {
		default: computeDefault(slugs),
		versions: ordered.map((slug) => ({ slug, label: slug })),
	};
}

export function renderRedirectHtml(defaultSlug) {
	const target = `${PAGES_PREFIX}/${defaultSlug}/`;
	return `<!doctype html>
<html lang="en">
  <head>
    <meta charset="utf-8" />
    <meta http-equiv="refresh" content="0; url=${target}" />
    <link rel="canonical" href="${target}" />
    <meta name="robots" content="noindex" />
    <title>Unswell documentation</title>
    <script>location.replace(${JSON.stringify(target)});</script>
  </head>
  <body>
    <p>Redirecting to the <a href="${target}">Unswell documentation</a>.</p>
  </body>
</html>
`;
}

function renderVersionsJson(index) {
	return `${JSON.stringify(index, null, 2)}\n`;
}

export function generate(dir) {
	const slugs = readdirSync(dir, { withFileTypes: true })
		.filter((entry) => entry.isDirectory() && isVersionFolder(entry.name))
		.map((entry) => entry.name);
	const index = buildIndex(slugs);
	writeFileSync(join(dir, 'versions.json'), renderVersionsJson(index));
	writeFileSync(join(dir, 'index.html'), renderRedirectHtml(index.default));
	return index;
}

function selftest() {
	const assert = (condition, message) => {
		if (!condition) throw new Error(message);
	};

	assert(isVersionFolder('edge'), 'edge is accepted');
	assert(isVersionFolder('v0.1.0-alpha.3'), 'a prerelease tag is accepted');
	assert(isVersionFolder('v1.2.0'), 'a release tag is accepted');
	assert(isVersionFolder('v1.2'), 'a minor tag is accepted');
	assert(!isVersionFolder('latest'), 'latest is not a version folder');
	assert(!isVersionFolder('_astro'), 'the asset folder is not a version folder');
	assert(!isVersionFolder('v1.0.0+build'), 'build metadata is refused');

	const order = (a, b) => compareSemver(parseSemver(a), parseSemver(b));
	assert(order('v1.0.0', 'v1.0.0-alpha.1') > 0, 'a release outranks its own prerelease');
	assert(order('v0.1.0-alpha.3', 'v0.1.0-alpha.2') > 0, 'numeric prerelease parts compare numerically');
	assert(order('v0.1.0-alpha.10', 'v0.1.0-alpha.9') > 0, 'prerelease numbers are not compared as text');
	assert(order('v0.1.0-beta.1', 'v0.1.0-alpha.9') > 0, 'text prerelease parts compare alphabetically');
	assert(order('v1.10.0', 'v1.2.0') > 0, 'minor numbers are not compared as text');

	assert(computeDefault(['edge']) === 'edge', 'edge is the default with no tag');
	assert(computeDefault([]) === 'edge', 'edge is the answer with nothing to choose from');
	assert(
		computeDefault(['edge', 'v0.1.0-alpha.3']) === 'v0.1.0-alpha.3',
		'a release outranks edge for the apex redirect'
	);
	assert(
		computeDefault(['edge', 'v0.1.0-alpha.2', 'v0.1.0-alpha.3']) === 'v0.1.0-alpha.3',
		'the newest release wins'
	);

	const index = buildIndex(['v0.1.0-alpha.2', 'edge', 'v0.1.0-alpha.3']);
	assert(index.default === 'v0.1.0-alpha.3', 'the default is the newest release');
	assert(
		index.versions.map((v) => v.slug).join(',') === 'edge,v0.1.0-alpha.3,v0.1.0-alpha.2',
		'edge leads, then releases newest first'
	);

	const tmp = mkdtempSync(join(tmpdir(), 'unswell-docs-versions-'));
	try {
		for (const version of ['edge', 'v0.1.0-alpha.3', '_astro']) {
			mkdirSync(join(tmp, version));
		}
		generate(tmp);
		const json1 = readFileSync(join(tmp, 'versions.json'), 'utf8');
		const html1 = readFileSync(join(tmp, 'index.html'), 'utf8');
		assert(html1.includes('/v0.1.0-alpha.3/'), 'the redirect targets the newest release');
		assert(!html1.includes('/edge/'), 'the redirect does not target edge beside a release');
		assert(!json1.includes('_astro'), 'a non-version folder is ignored');
		generate(tmp);
		assert(json1 === readFileSync(join(tmp, 'versions.json'), 'utf8'), 'versions.json is idempotent');
		assert(html1 === readFileSync(join(tmp, 'index.html'), 'utf8'), 'index.html is idempotent');
		console.log('gen-versions.mjs --selftest: OK');
	} finally {
		rmSync(tmp, { recursive: true, force: true });
	}
}

function main() {
	const arg = process.argv[2];
	if (arg === '--selftest') {
		selftest();
		return;
	}
	if (!arg) {
		console.error('usage: node scripts/gen-versions.mjs <site-dir> | --selftest');
		process.exitCode = 2;
		return;
	}
	if (!existsSync(arg)) {
		console.error(`error: directory not found: ${arg}`);
		process.exitCode = 2;
		return;
	}
	const index = generate(arg);
	console.log(`wrote ${join(arg, 'versions.json')} (default=${index.default})`);
}

if (process.argv[1] && process.argv[1].endsWith('gen-versions.mjs')) {
	main();
}
