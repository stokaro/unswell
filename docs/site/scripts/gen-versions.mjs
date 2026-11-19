#!/usr/bin/env node
// Writes the files that describe the assembled Pages root: versions.json, which
// the version picker reads, index.html, the apex redirect, and the addresses
// that do not depend on a release.
//
// The assembly holds one directory per documentation version. This script reads
// the directory names rather than a list kept by hand, so a version that failed
// to build is a version the picker does not offer and the redirect cannot land
// on.
//
// A reader copies the address of a page with its version in it, and the site
// keeps only the latest releases. So every page of the default version also
// gets an address without a version, a stub that sends the reader to the page
// of the latest release, and 404.html sends an address of a release that is no
// longer built to that stub.
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

// PAGE_RE is what a page path may hold. The paths come from folder names of the
// build, and a stub writes its path into HTML, so anything else is refused.
const PAGE_RE = /^[A-Za-z0-9._-]+(?:\/[A-Za-z0-9._-]+)*$/;

// pagePaths lists the pages of one version: every folder below it that holds
// an index.html, as a path without a slash at either end. The front page of
// the version is the apex redirect's business. Folders of assets and of the
// search index hold no page.
export function pagePaths(dir, slug) {
	const found = [];
	const walk = (prefix) => {
		for (const entry of readdirSync(join(dir, slug, prefix), { withFileTypes: true })) {
			if (!entry.isDirectory() || entry.name.startsWith('_') || entry.name === 'pagefind') continue;
			const relative = prefix === '' ? entry.name : `${prefix}/${entry.name}`;
			if (existsSync(join(dir, slug, relative, 'index.html'))) found.push(relative);
			walk(relative);
		}
	};
	if (existsSync(join(dir, slug))) walk('');
	return found.sort();
}

// renderPageRedirectHtml is the stub at the address of a page without a
// version. The script keeps the query and the fragment of the link, which a
// refresh header cannot do, so the header is there only for a reader without
// JavaScript.
export function renderPageRedirectHtml(defaultSlug, page) {
	if (!PAGE_RE.test(page)) throw new Error(`page path ${JSON.stringify(page)} cannot be published as an address`);
	const target = `${PAGES_PREFIX}/${defaultSlug}/${page}/`;
	return `<!doctype html>
<html lang="en">
  <head>
    <meta charset="utf-8" />
    <link rel="canonical" href="${target}" />
    <meta name="robots" content="noindex" />
    <title>Unswell documentation</title>
    <script>location.replace(${JSON.stringify(target)} + location.search + location.hash);</script>
    <noscript><meta http-equiv="refresh" content="0; url=${target}" /></noscript>
  </head>
  <body>
    <p>This page is at <a href="${target}">${target}</a> in the latest release.</p>
  </body>
</html>
`;
}

// versionlessPath returns the address without its version for a path whose
// first part names a version, and null for any other path. 404.html runs the
// same function in the browser: Pages serves that page for an address of a
// release that is no longer built, and the address without the version is the
// stub of the same page.
export function versionlessPath(pathname, versionSource = VERSION_RE.source) {
	const parts = pathname.split('/').filter(Boolean);
	if (parts.length === 0 || !(parts[0] === 'edge' || new RegExp(versionSource).test(parts[0]))) return null;
	const rest = parts.slice(1).join('/');
	return `/${rest === '' ? '' : `${rest}/`}`;
}

// renderNotFoundHtml is the page that Pages serves for an address the site
// does not hold. An address under a version goes to the same page without the
// version. That address either is a stub or is not found again, and then its
// first part is no version, so the script runs once at most.
export function renderNotFoundHtml() {
	return `<!doctype html>
<html lang="en">
  <head>
    <meta charset="utf-8" />
    <meta name="robots" content="noindex" />
    <title>Page not found | Unswell documentation</title>
    <script>
      var versionlessPath = ${versionlessPath.toString()};
      var target = versionlessPath(location.pathname, ${JSON.stringify(VERSION_RE.source)});
      if (target !== null) location.replace(target + location.search + location.hash);
    </script>
  </head>
  <body>
    <p>This address holds no page. The documentation starts at <a href="${PAGES_PREFIX}/">the front page</a>.</p>
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
	writeFileSync(join(dir, '404.html'), renderNotFoundHtml());
	const pages = pagePaths(dir, index.default);
	for (const page of pages) {
		// A page with the name of a version would shadow that version's folder.
		if (isVersionFolder(page.split('/')[0])) {
			throw new Error(`page ${page} of ${index.default} has the name of a version`);
		}
		mkdirSync(join(dir, page), { recursive: true });
		writeFileSync(join(dir, page, 'index.html'), renderPageRedirectHtml(index.default, page));
	}
	return { ...index, pages };
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

		// Every page of the default version gets an address without a version.
		for (const page of ['research-report', 'guides/install']) {
			mkdirSync(join(tmp, 'v0.1.0-alpha.3', page), { recursive: true });
			writeFileSync(join(tmp, 'v0.1.0-alpha.3', page, 'index.html'), 'page');
		}
		mkdirSync(join(tmp, 'v0.1.0-alpha.3', '_astro', 'chunk'), { recursive: true });
		writeFileSync(join(tmp, 'v0.1.0-alpha.3', '_astro', 'chunk', 'index.html'), 'asset');
		mkdirSync(join(tmp, 'v0.1.0-alpha.3', 'pagefind'), { recursive: true });
		writeFileSync(join(tmp, 'v0.1.0-alpha.3', 'pagefind', 'index.html'), 'search');
		mkdirSync(join(tmp, 'edge', 'only-on-edge'), { recursive: true });
		writeFileSync(join(tmp, 'edge', 'only-on-edge', 'index.html'), 'page');
		const first = generate(tmp);
		assert(first.pages.join(',') === 'guides/install,research-report', 'the pages of the default version are listed');
		const stub = readFileSync(join(tmp, 'research-report', 'index.html'), 'utf8');
		assert(stub.includes('href="/v0.1.0-alpha.3/research-report/"'), 'a stub names the page of the default version');
		assert(stub.includes('+ location.search + location.hash'), 'a stub keeps the query and the fragment');
		assert(stub.includes('noindex'), 'a stub stays out of a search index');
		assert(existsSync(join(tmp, 'guides', 'install', 'index.html')), 'a nested page gets a stub');
		assert(!existsSync(join(tmp, '_astro', 'chunk', 'index.html')), 'an asset folder gets no stub');
		assert(!existsSync(join(tmp, 'pagefind', 'index.html')), 'the search index gets no stub');
		assert(!existsSync(join(tmp, 'only-on-edge')), 'a page that only edge holds gets no stub');
		const second = generate(tmp);
		assert(second.pages.join(',') === first.pages.join(','), 'the stubs of an earlier run are not read as pages or versions');
		assert(stub === readFileSync(join(tmp, 'research-report', 'index.html'), 'utf8'), 'a stub is idempotent');
		assert(json1 === readFileSync(join(tmp, 'versions.json'), 'utf8'), 'the stubs do not change versions.json');

		let refused = false;
		try {
			renderPageRedirectHtml('v0.1.0-alpha.3', 'a"><script>');
		} catch {
			refused = true;
		}
		assert(refused, 'a page path that would break out of the markup is refused');
		mkdirSync(join(tmp, 'v0.1.0-alpha.3', 'v9.9.9'), { recursive: true });
		writeFileSync(join(tmp, 'v0.1.0-alpha.3', 'v9.9.9', 'index.html'), 'page');
		refused = false;
		try {
			generate(tmp);
		} catch {
			refused = true;
		}
		assert(refused, 'a page with the name of a version is refused');

		// 404.html sends an address under a version to the same page without it.
		assert(versionlessPath('/v0.1.0-alpha.2/research-report/') === '/research-report/', 'a removed release goes to the stub');
		assert(versionlessPath('/v0.1.0-alpha.2/guides/install') === '/guides/install/', 'a nested page keeps its path');
		assert(versionlessPath('/v0.1.0-alpha.2/') === '/', 'a version alone goes to the apex');
		assert(versionlessPath('/edge/gone/') === '/gone/', 'edge is a version too');
		assert(versionlessPath('/research-report/') === null, 'an address without a version is left alone, so nothing loops');
		assert(versionlessPath('/') === null, 'the apex is left alone');
		assert(versionlessPath('/latest/x/') === null, 'a word that is no version is left alone');
		const notFound = readFileSync(join(tmp, '404.html'), 'utf8');
		assert(notFound.includes('versionlessPath(location.pathname'), '404.html runs the same function');
		assert(notFound.includes(JSON.stringify(VERSION_RE.source)), '404.html holds the pattern of a version');
		const inPage = new Function(`${/<script>([\s\S]*?)<\/script>/.exec(notFound)[1].split('var target')[0]}; return versionlessPath;`)();
		assert(inPage('/v0.1.0-alpha.2/research-report/', VERSION_RE.source) === '/research-report/', 'the function runs as the page holds it');
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
	console.log(`wrote ${join(arg, 'versions.json')} (default=${index.default}) and ${index.pages.length} addresses without a version`);
}

if (process.argv[1] && process.argv[1].endsWith('gen-versions.mjs')) {
	main();
}
