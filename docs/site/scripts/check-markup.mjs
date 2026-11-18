#!/usr/bin/env node
// Refuses a paragraph that holds a block.
//
// A paragraph cannot contain another paragraph, a div, a list, or a table. A
// browser does not complain about one that does: it closes the paragraph at the
// inner block, so a box that a component drew around a slot stays empty and its
// text falls out under it. The source looks right, the build passes, and only a
// reader sees the empty box. This gate reads the built pages and finds the
// place.
import { existsSync, readdirSync, readFileSync, mkdtempSync, mkdirSync, rmSync, writeFileSync } from 'node:fs';
import { tmpdir } from 'node:os';
import { dirname, join, posix, resolve } from 'node:path';
import { fileURLToPath } from 'node:url';

const scriptDir = dirname(fileURLToPath(import.meta.url));
const defaultDist = join(scriptDir, '..', 'dist');

// The elements whose start tag closes an open paragraph, by the HTML standard.
const CLOSES_PARAGRAPH = new Set([
	'address', 'article', 'aside', 'blockquote', 'details', 'div', 'dl', 'fieldset', 'figcaption', 'figure', 'footer',
	'form', 'h1', 'h2', 'h3', 'h4', 'h5', 'h6', 'header', 'hgroup', 'hr', 'main', 'menu', 'nav', 'ol', 'p', 'pre',
	'section', 'table', 'ul',
]);

// Text that holds no markup of the page: scripts, styles, comments, and what an
// SVG or a template carries.
const OPAQUE_RE = /<!--[\s\S]*?-->|<(script|style|template|svg)\b[\s\S]*?<\/\1\s*>/gi;
const TAG_RE = /<(\/?)([a-zA-Z][a-zA-Z0-9]*)\b[^>]*>/g;

// brokenParagraphs returns, for one document, each block that opens inside an
// open paragraph, with the paragraph's own start tag.
export function brokenParagraphs(html) {
	const found = [];
	const text = html.replace(OPAQUE_RE, '');
	let open = null;
	for (const match of text.matchAll(TAG_RE)) {
		const [tag, slash, rawName] = match;
		const name = rawName.toLowerCase();
		if (slash) {
			if (name === 'p') {
				open = null;
			}
			continue;
		}
		if (open !== null && CLOSES_PARAGRAPH.has(name)) {
			found.push({ paragraph: open, block: tag });
			open = null;
		}
		if (name === 'p') {
			open = tag;
		}
	}
	return found;
}

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

export function findBroken(dist) {
	const broken = [];
	for (const file of htmlFiles(dist)) {
		for (const fault of brokenParagraphs(readFileSync(join(dist, file), 'utf8'))) {
			broken.push({ file, ...fault });
		}
	}
	return broken;
}

function assert(condition, message) {
	if (!condition) {
		throw new Error(`selftest failed: ${message}`);
	}
}

function selftest() {
	const nested = brokenParagraphs('<p class="note"><p>Text of the note.</p></p>');
	assert(nested.length === 1 && nested[0].paragraph === '<p class="note">' && nested[0].block === '<p>', 'a paragraph inside a paragraph is found');
	assert(brokenParagraphs('<p class="body"><div>A block.</div></p>').length === 1, 'a div inside a paragraph is found');
	assert(brokenParagraphs('<p>One <ul><li>item</li></ul></p>').length === 1, 'a list inside a paragraph is found');
	assert(brokenParagraphs('<div class="note"><p>Text.</p><p>More text.</p></div>').length === 0, 'paragraphs inside a div pass');
	assert(brokenParagraphs('<p>Text with <a href="/x/">a link</a>, <code>code</code>, and <strong>weight</strong>.</p><p>Next.</p>').length === 0, 'inline elements in a paragraph pass');
	assert(brokenParagraphs('<p>Text.</p><script>const html = "<p><p>";</script><!-- <p><div> --><p>More.</p>').length === 0, 'markup inside a script or a comment is not read');
	assert(brokenParagraphs('<P CLASS="x"><DIV>upper case</DIV></P>').length === 1, 'tag names are read without regard to case');

	const tmp = mkdtempSync(join(tmpdir(), 'unswell-markup-'));
	try {
		mkdirSync(join(tmp, 'page'), { recursive: true });
		writeFileSync(join(tmp, 'index.html'), '<div><p>Good.</p></div>');
		writeFileSync(join(tmp, 'page', 'index.html'), '<p class="u-note"><p>Lost.</p></p>');
		writeFileSync(join(tmp, 'page', 'data.json'), '"<p><p>"');
		const broken = findBroken(tmp);
		assert(broken.length === 1 && broken[0].file === 'page/index.html', 'only the page with the broken paragraph is named');
	} finally {
		rmSync(tmp, { recursive: true, force: true });
	}
	console.log('check-markup.mjs --selftest: OK');
}

function main() {
	if (process.argv.includes('--selftest')) {
		selftest();
		return;
	}
	const index = process.argv.indexOf('--dist');
	const dist = index === -1 ? defaultDist : resolve(process.argv[index + 1]);
	if (!existsSync(dist)) {
		console.error(`check-markup: no build at ${dist}`);
		process.exitCode = 2;
		return;
	}
	const broken = findBroken(dist);
	for (const { file, paragraph, block } of broken) {
		console.error(`check-markup: ${file}: ${block} opens inside ${paragraph}; a browser closes the paragraph there`);
	}
	if (broken.length > 0) {
		process.exitCode = 1;
		return;
	}
	console.log('check-markup: no paragraph of the built pages holds a block');
}

if (process.argv[1] && process.argv[1].endsWith('check-markup.mjs')) {
	main();
}
