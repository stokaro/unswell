// The geometry of the figures of the research report.
//
// A figure arrives as data: panels, each with a form, a unit, series, and rows of
// values that carry a number and the words for it. This module turns a figure
// into primitives with coordinates and class names, and Figure.astro writes them
// out as SVG. Nothing here knows a color: the classes are painted by the style
// sheet, so one drawing serves both themes.
//
// Text is set in the monospaced face. Its letters are all one width, which is the
// only way a static build can know how wide a label is before a browser draws it.

export const WIDTH = 720;
const CHAR = 6.7; // width of one letter of 11px IBM Plex Mono
const FONT = 11;
const PAD = 8;
const BAR = 14; // thickness of one bar
const BAR_GAP = 3; // between the bars of one row
const ROW_GAP = 16; // between rows
const RADIUS = 4;
const DOT = 5;

const plain = (label) => label.replaceAll('`', '');
const widthOf = (text) => plain(String(text)).length * CHAR;

// ticks returns round positions on a scale from min to max.
export function ticks(min, max, unit) {
	const steps = { share: max <= 0.35 ? 0.1 : 0.25, auc: 0.25, index: 25 };
	let step = steps[unit];
	if (!step) {
		const rough = (max - min) / 4;
		const power = 10 ** Math.floor(Math.log10(rough));
		step = [1, 2, 5, 10].map((m) => m * power).find((s) => s >= rough);
	}
	const found = [];
	for (let value = Math.ceil(min / step) * step; value <= max + step / 1000; value += step) {
		found.push(Number(value.toFixed(6)));
	}
	return found;
}

export function tickLabel(value, unit) {
	if (unit === 'share') return `${Math.round(value * 100)}%`;
	if (unit === 'auc') return value.toFixed(2).replace(/0$/, '').replace(/\.$/, '.0');
	return String(value);
}

function limits(panel) {
	const values = panel.rows.flatMap((row) =>
		panel.stacked
			? [Object.values(row.values).reduce((sum, cell) => sum + cell.value, 0)]
			: Object.values(row.values).map((cell) => cell.value)
	);
	const min = panel.min ?? 0;
	let max = panel.max ?? Math.max(...values);
	if (panel.max === undefined) {
		const step = ticks(min, max, panel.unit)[1] ?? max;
		max = Math.ceil(max / step) * step;
	}
	return { min, max };
}

// bar is a path with a rounded data end and a square foot at the baseline.
function barPath(x, y, length, thick, vertical = false) {
	const r = Math.min(RADIUS, length);
	if (vertical) {
		// x, y is the foot of the column at its left; it grows up.
		return `M${x},${y} V${y - length + r} Q${x},${y - length} ${x + r},${y - length} H${x + thick - r} Q${x + thick},${y - length} ${x + thick},${y - length + r} V${y} Z`;
	}
	return `M${x},${y} H${x + length - r} Q${x + length},${y} ${x + length},${y + r} V${y + thick - r} Q${x + length},${y + thick} ${x + length - r},${y + thick} H${x} Z`;
}

function tip(row, series, cell) {
	const detail = cell.detail ? ` (${cell.detail})` : '';
	return `${plain(row.label)}, ${plain(series.label)}: ${cell.text}${detail}`;
}

function axis(out, panel, scale, left, right, top, bottom) {
	for (const value of ticks(scale.min, scale.max, panel.unit)) {
		const x = left + ((value - scale.min) / (scale.max - scale.min)) * (right - left);
		out.lines.push({ x1: x, y1: top, x2: x, y2: bottom, class: 'u-fig-grid' });
		out.texts.push({ x, y: bottom + FONT + 4, text: tickLabel(value, panel.unit), anchor: 'middle', class: 'u-fig-muted' });
	}
	if (panel.reference) {
		const x = left + ((panel.reference.value - scale.min) / (scale.max - scale.min)) * (right - left);
		out.lines.push({ x1: x, y1: top - 4, x2: x, y2: bottom, class: 'u-fig-ref' });
		out.texts.push({ x: x + 4, y: top - 6, text: panel.reference.label, anchor: 'start', class: 'u-fig-muted' });
	}
}

function bars(panel, top, shared = {}) {
	const out = { rects: [], paths: [], lines: [], texts: [], dots: [] };
	const scale = limits(panel);
	// The panels of one figure start their bars at one edge, so they can be read against each other.
	const labelWidth = (shared.labelWidth ?? Math.max(...panel.rows.map((row) => widthOf(row.label)))) + PAD;
	const valueWidth = Math.max(
		...panel.rows.flatMap((row) => (row.total ? [widthOf(row.total)] : Object.values(row.values).map((cell) => widthOf(cell.text))))
	);
	const left = labelWidth + PAD;
	const right = WIDTH - valueWidth - 2 * PAD;
	const perRow = panel.stacked ? 1 : panel.series.length;
	const rowHeight = perRow * BAR + (perRow - 1) * BAR_GAP;
	const plotTop = top + (panel.reference ? FONT + 8 : 0);
	const bottom = plotTop + panel.rows.length * (rowHeight + ROW_GAP) - ROW_GAP + 6;
	axis(out, panel, scale, left, right, plotTop, bottom);
	const length = (value) => ((value - scale.min) / (scale.max - scale.min)) * (right - left);
	// A label that would stand across the reference line moves to the right of it.
	const mark = panel.reference ? left + length(panel.reference.value) : null;
	const clear = (x, text) => (mark !== null && x <= mark + 2 && x + widthOf(text) >= mark - 2 ? mark + 6 : x);
	panel.rows.forEach((row, index) => {
		const y = plotTop + index * (rowHeight + ROW_GAP);
		out.texts.push({ x: left - PAD, y: y + rowHeight / 2 + FONT / 2 - 2, text: plain(row.label), anchor: 'end', class: 'u-fig-text' });
		if (panel.stacked) {
			let x = left;
			panel.series.forEach((series, slot) => {
				const cell = row.values[series.key];
				const w = length(cell.value);
				if (w <= 0) return;
				const last = panel.series.slice(slot + 1).every((later) => row.values[later.key].value === 0);
				const gap = slot > 0 ? 2 : 0;
				const d = last ? barPath(x + gap, y, w - gap, BAR) : `M${x + gap},${y} H${x + w} V${y + BAR} H${x + gap} Z`;
				out.paths.push({ d, class: `u-fig-s${slot + 1}`, hatched: Boolean(series.hatched), title: tip(row, series, cell) });
				x += w;
			});
			out.texts.push({ x: x + 6, y: y + BAR / 2 + FONT / 2 - 2, text: row.total, anchor: 'start', class: 'u-fig-text' });
			return;
		}
		panel.series.forEach((series, slot) => {
			const cell = row.values[series.key];
			const by = y + slot * (BAR + BAR_GAP);
			const w = length(cell.value);
			if (w > 0) {
				out.paths.push({ d: barPath(left, by, w, BAR), class: `u-fig-s${slot + 1}`, title: tip(row, series, cell) });
			} else {
				// A value of zero draws no bar; a short stroke at the baseline says that the row was measured.
				out.lines.push({ x1: left, y1: by + 2, x2: left, y2: by + BAR - 2, class: `u-fig-zero u-fig-stroke-s${slot + 1}`, title: tip(row, series, cell) });
			}
			out.texts.push({ x: clear(left + w + 6, cell.text), y: by + BAR / 2 + FONT / 2 - 2, text: cell.text, anchor: 'start', class: 'u-fig-text' });
		});
	});
	return { ...out, height: bottom + FONT + 10 - top };
}

function dumbbell(panel, top) {
	const out = { rects: [], paths: [], lines: [], texts: [], dots: [] };
	const scale = limits(panel);
	const labelWidth = Math.max(...panel.rows.map((row) => widthOf(row.label))) + PAD;
	const left = labelWidth + PAD + 30;
	const right = WIDTH - 44;
	const rowHeight = 22;
	const plotTop = top + (panel.reference ? FONT + 8 : 0);
	const bottom = plotTop + panel.rows.length * rowHeight + 4;
	axis(out, panel, scale, left, right, plotTop, bottom);
	const at = (value) => left + ((value - scale.min) / (scale.max - scale.min)) * (right - left);
	panel.rows.forEach((row, index) => {
		const y = plotTop + index * rowHeight + rowHeight / 2;
		out.texts.push({ x: labelWidth, y: y + FONT / 2 - 2, text: plain(row.label), anchor: 'end', class: 'u-fig-text' });
		const [first, second] = panel.series.map((series) => row.values[series.key]);
		const [x1, x2] = [at(first.value), at(second.value)];
		out.lines.push({ x1, y1: y, x2, y2: y, class: 'u-fig-link' });
		panel.series.forEach((series, slot) => {
			const cell = row.values[series.key];
			const x = at(cell.value);
			out.dots.push({ cx: x, cy: y, r: DOT, class: `u-fig-s${slot + 1}`, title: tip(row, series, cell) });
			const outward = (slot === 0) === (x1 <= x2) ? -1 : 1;
			out.texts.push({ x: x + outward * (DOT + 5), y: y + FONT / 2 - 2, text: cell.text, anchor: outward < 0 ? 'end' : 'start', class: 'u-fig-text' });
		});
	});
	return { ...out, height: bottom + FONT + 10 - top };
}

// releases shortens the tags of one week: alpha.5 to alpha.15 becomes "alpha.5–15".
export function releases(tags) {
	if (tags.length === 0) return '';
	const numbers = tags.map((tag) => Number(tag.replace('alpha.', '')));
	return numbers.length === 1 ? `alpha.${numbers[0]}` : `alpha.${numbers[0]}–${numbers[numbers.length - 1]}`;
}

function columns(panel, top) {
	const out = { rects: [], paths: [], lines: [], texts: [], dots: [] };
	const scale = limits(panel);
	const left = 44;
	const right = WIDTH - PAD;
	const plotTop = top + FONT + 6;
	const plotHeight = 180;
	const bottom = plotTop + plotHeight;
	const series = panel.series[0];
	for (const value of ticks(scale.min, scale.max, panel.unit)) {
		const y = bottom - ((value - scale.min) / (scale.max - scale.min)) * plotHeight;
		out.lines.push({ x1: left, y1: y, x2: right, y2: y, class: 'u-fig-grid' });
		out.texts.push({ x: left - 6, y: y + FONT / 2 - 2, text: String(value), anchor: 'end', class: 'u-fig-muted' });
	}
	const slot = (right - left) / panel.rows.length;
	const thick = Math.min(24, slot - 8);
	const band = bottom + 2 * FONT + 14; // the row of release marks under the dates
	let level = 1;
	panel.rows.forEach((row, index) => {
		const cell = row.values[series.key];
		const x = left + index * slot + (slot - thick) / 2;
		const height = ((cell.value - scale.min) / (scale.max - scale.min)) * plotHeight;
		out.paths.push({ d: barPath(x, bottom, height, thick, true), class: 'u-fig-s1', title: `Week of ${row.label}: ${cell.text} commits` });
		out.texts.push({ x: x + thick / 2, y: bottom - height - 5, text: cell.text, anchor: 'middle', class: 'u-fig-text' });
		out.texts.push({ x: x + thick / 2, y: bottom + FONT + 4, text: row.label, anchor: 'middle', class: 'u-fig-muted' });
		if (!row.events?.length) {
			level = 1;
			return;
		}
		// Two weeks side by side both have releases: the second label goes one line lower, so they do not meet.
		level = level === 0 ? 1 : 0;
		out.dots.push({ cx: x + thick / 2, cy: band, r: 4, class: 'u-fig-s2', title: `Releases: ${row.events.join(', ')}` });
		out.texts.push({ x: x + thick / 2, y: band + FONT + 8 + level * (FONT + 4), text: releases(row.events), anchor: 'middle', class: 'u-fig-text' });
	});
	out.texts.push({ x: 0, y: band + 4, text: 'Releases', anchor: 'start', class: 'u-fig-muted' });
	return { ...out, height: band + 2 * FONT + 22 - top };
}

const FORMS = { bars, dumbbell, columns };

// sameSeries reports whether every panel of a figure has the series of the first.
function sameSeries(figure) {
	const first = JSON.stringify(figure.panels[0].series);
	return figure.panels.every((panel) => JSON.stringify(panel.series) === first);
}

function legend(series, top) {
	const items = [];
	let x = 0;
	series.forEach((entry, slot) => {
		items.push({ x, y: top, slot: slot + 1, hatched: Boolean(entry.hatched), text: plain(entry.label) });
		x += 14 + 6 + widthOf(entry.label) + 22;
	});
	return items;
}

// layout turns a figure into primitives and the height of its drawing.
export function layout(figure) {
	const shared = sameSeries(figure);
	const drawn = { width: WIDTH, legends: [], titles: [], rects: [], paths: [], lines: [], texts: [], dots: [] };
	let y = 4;
	const needsLegend = (panel) => panel.series.length > 1;
	if (shared && needsLegend(figure.panels[0])) {
		drawn.legends.push(...legend(figure.panels[0].series, y));
		y += 26;
	}
	const labelWidth = Math.max(...figure.panels.flatMap((panel) => panel.rows.map((row) => widthOf(row.label))));
	for (const panel of figure.panels) {
		const form = FORMS[panel.form];
		if (!form) throw new Error(`unknown form of a panel: ${panel.form}`);
		if (panel.title) {
			drawn.titles.push({ x: 0, y: y + FONT, text: panel.title });
			y += FONT + 12;
		}
		if (!shared && needsLegend(panel)) {
			drawn.legends.push(...legend(panel.series, y));
			y += 26;
		}
		const part = form(panel, y, { labelWidth });
		for (const key of ['rects', 'paths', 'lines', 'texts', 'dots']) drawn[key].push(...part[key]);
		y += part.height + 14;
	}
	drawn.height = Math.ceil(y);
	return drawn;
}

// outside lists the texts of a drawing that leave it on the left or the right.
export function outside(drawn) {
	return drawn.texts.filter((text) => {
		const width = widthOf(text.text);
		const start = text.anchor === 'end' ? text.x - width : text.anchor === 'middle' ? text.x - width / 2 : text.x + (text.shift ?? 0);
		return start < -1 || start + width > drawn.width + 1;
	});
}
