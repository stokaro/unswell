#!/usr/bin/env node
// Refuses a link to a file of this repository that is not in the tree.
//
// Pages cite records, decisions, and scripts by their address on GitHub, on the
// main branch. The other link gate does not follow an address with a scheme, so
// a renamed record would leave a dead link that no check sees. This one reads
// the built pages and looks each such path up in the working tree that the site
// was built from.
//
// It also draws every figure of the research report from its data, and refuses
// a label that leaves the drawing or a bar that is not as long as its number.
import { existsSync, readdirSync, readFileSync, mkdtempSync, mkdirSync, rmSync, writeFileSync } from 'node:fs';
import { tmpdir } from 'node:os';
import { dirname, join, posix, resolve } from 'node:path';
import { fileURLToPath } from 'node:url';

import { RepositoryURL } from '../src/lib/docs-origin.mjs';
import { layout, outside, WIDTH } from '../src/lib/chart.mjs';

const scriptDir = dirname(fileURLToPath(import.meta.url));
const siteRoot = join(scriptDir, '..');
const defaultDist = join(siteRoot, 'dist');
const defaultTree = resolve(siteRoot, '..', '..');

const HREF_RE = /href="([^"]*)"/g;

function htmlFiles(root, prefix = '') {
	const found = [];
	for (const entry of readdirSync(join(root, prefix), { withFileTypes: true })) {
		const relative = posix.join(prefix, entry.name);
		if (entry.isDirectory()) {
			found.push(...htmlFiles(root, relative));
		} else if (entry.name.endsWith('.html')) {
			found.push(relative);
		}
	}
	return found;
}

// repositoryPaths returns the paths on the main branch that one page links to.
export function repositoryPaths(html, repository = RepositoryURL) {
	const paths = new Set();
	for (const match of html.matchAll(HREF_RE)) {
		const href = match[1].replaceAll('&amp;', '&');
		for (const kind of ['blob', 'tree']) {
			const prefix = `${repository}/${kind}/main/`;
			if (href.startsWith(prefix)) {
				paths.add(decodeURIComponent(href.slice(prefix.length).split('#')[0].split('?')[0]));
			}
		}
	}
	return [...paths];
}

// findMissing lists, for each page, the linked paths that the tree lacks.
export function findMissing(dist, tree, repository = RepositoryURL) {
	const missing = [];
	for (const file of htmlFiles(dist)) {
		for (const path of repositoryPaths(readFileSync(join(dist, file), 'utf8'), repository)) {
			if (path.includes('..') || !existsSync(join(tree, path))) {
				missing.push({ file, path });
			}
		}
	}
	return missing;
}

// figureFaults draws every figure of a report and lists what is wrong with a drawing.
export function figureFaults(report) {
	const faults = [];
	for (const [name, figure] of Object.entries(report.figures)) {
		const drawn = layout(figure);
		for (const text of outside(drawn)) {
			faults.push(`figure ${name}: the label "${text.text}" leaves the drawing`);
		}
		if (drawn.width !== WIDTH || !(drawn.height > 0)) {
			faults.push(`figure ${name}: no size`);
		}
		const marks = drawn.paths.length + drawn.dots.length + drawn.lines.filter((line) => line.title).length;
		const values = figure.panels.reduce(
			(sum, panel) => sum + panel.rows.reduce((cells, row) => cells + Object.values(row.values).filter((cell) => !panel.stacked || cell.value > 0).length, 0),
			0
		);
		const events = figure.panels.reduce((sum, panel) => sum + panel.rows.filter((row) => row.events?.length).length, 0);
		if (marks !== values + events) {
			faults.push(`figure ${name}: ${marks} marks for ${values + events} values`);
		}
	}
	return faults;
}

function assert(condition, message) {
	if (!condition) {
		throw new Error(`selftest failed: ${message}`);
	}
}

function selftest() {
	const repository = 'https://github.com/example/project';
	const html = [
		`<a href="${repository}/blob/main/docs/kept.md#part">kept</a>`,
		`<a href="${repository}/tree/main/research">a directory</a>`,
		`<a href="${repository}/blob/main/docs/gone.md">gone</a>`,
		`<a href="${repository}/issues/7">an issue</a>`,
		`<a href="https://example.org/blob/main/docs/gone.md">another site</a>`,
		`<a href="${repository}/blob/v1/docs/gone.md">a tag</a>`,
	].join('\n');
	assert(
		repositoryPaths(html, repository).join(',') === 'docs/kept.md,research,docs/gone.md',
		'only links to the main branch of this repository are paths of the tree'
	);
	const tmp = mkdtempSync(join(tmpdir(), 'unswell-repo-links-'));
	try {
		const dist = join(tmp, 'dist');
		const tree = join(tmp, 'tree');
		mkdirSync(join(dist, 'page'), { recursive: true });
		mkdirSync(join(tree, 'docs'), { recursive: true });
		mkdirSync(join(tree, 'research'), { recursive: true });
		writeFileSync(join(tree, 'docs', 'kept.md'), 'kept\n');
		writeFileSync(join(dist, 'page', 'index.html'), html);
		const missing = findMissing(dist, tree, repository);
		assert(missing.length === 1 && missing[0].path === 'docs/gone.md', 'a link to a file that the tree lacks is found');
		writeFileSync(join(tree, 'docs', 'gone.md'), 'back\n');
		assert(findMissing(dist, tree, repository).length === 0, 'a tree that holds every linked file passes');
	} finally {
		rmSync(tmp, { recursive: true, force: true });
	}

	const cell = (value, text) => ({ value, text });
	const figure = (label, value, max) => ({
		title: 'A figure',
		panels: [{ form: 'bars', unit: 'share', max, series: [{ key: 'a', label: 'A' }], rows: [{ label, values: { a: cell(value, '1 of 2') } }] }],
	});
	assert(figureFaults({ figures: { good: figure('A row', 0.5, 1) } }).length === 0, 'a figure inside its drawing passes');
	const long = 'A label that is far too long to stand beside a bar of a figure that is seven hundred pixels wide, by any measure of it';
	assert(
		figureFaults({ figures: { wide: figure(long, 0.5, 1) } }).some((fault) => fault.includes('leaves the drawing')),
		'a label wider than the drawing is found'
	);
	const half = layout(figure('A row', 0.5, 1)).paths[0].d;
	const full = layout(figure('A row', 1, 1)).paths[0].d;
	const reach = (d) => Number(d.match(/H([\d.]+)/)[1]) + 4 - Number(d.match(/^M([\d.]+)/)[1]);
	assert(Math.abs(reach(full) - 2 * reach(half)) < 0.01, 'a bar of twice the value is twice as long');
	console.log('check-repo-links.mjs --selftest: OK');
}

function main() {
	if (process.argv.includes('--selftest')) {
		selftest();
		return;
	}
	const option = (name, fallback) => {
		const index = process.argv.indexOf(name);
		return index === -1 ? fallback : resolve(process.argv[index + 1]);
	};
	const dist = option('--dist', defaultDist);
	const tree = option('--tree', defaultTree);
	if (!existsSync(dist)) {
		console.error(`check-repo-links: no build at ${dist}`);
		process.exitCode = 2;
		return;
	}
	const missing = findMissing(dist, tree);
	for (const { file, path } of missing) {
		console.error(`check-repo-links: ${file} links ${path}, which the repository does not hold`);
	}
	const report = JSON.parse(readFileSync(join(siteRoot, 'src', 'data', 'research-report.json'), 'utf8'));
	const faults = figureFaults(report);
	for (const fault of faults) {
		console.error(`check-repo-links: ${fault}`);
	}
	if (missing.length > 0 || faults.length > 0) {
		process.exitCode = 1;
		return;
	}
	console.log('check-repo-links: every linked file of the repository exists, and every figure fits its drawing');
}

if (process.argv[1] && process.argv[1].endsWith('check-repo-links.mjs')) {
	main();
}
