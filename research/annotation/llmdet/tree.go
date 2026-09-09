package llmdet

import "fmt"

func validateTree(tree TreeSpec, features, classes int) error {
	if tree.Class < 0 || tree.Class >= classes || tree.Comparison != "numeric-le" ||
		len(tree.Leaves) < 1 || len(tree.Leaves) > maxLeaves || len(tree.Splits) != len(tree.Leaves)-1 {
		return fmt.Errorf("invalid tree dimensions or comparison")
	}
	if err := validateTreeValues(tree, features); err != nil {
		return err
	}
	return validateTreeGraph(tree)
}

func validateTreeValues(tree TreeSpec, features int) error {
	for _, leaf := range tree.Leaves {
		if !finite(leaf) {
			return fmt.Errorf("nonfinite leaf")
		}
	}
	for _, split := range tree.Splits {
		if split.Feature < 0 || split.Feature >= features || !finite(split.Threshold) {
			return fmt.Errorf("invalid split feature or threshold")
		}
		if !validChild(split.Left, tree) || !validChild(split.Right, tree) {
			return fmt.Errorf("invalid split child")
		}
	}
	return nil
}

func validateTreeGraph(tree TreeSpec) error {
	if len(tree.Splits) == 0 {
		return nil
	}
	walk := treeWalk{tree: tree, splits: make([]bool, len(tree.Splits)), leaves: make([]bool, len(tree.Leaves))}
	if err := walk.visit(0, 0); err != nil {
		return err
	}
	for _, visited := range append(walk.splits, walk.leaves...) {
		if !visited {
			return fmt.Errorf("unreachable tree node")
		}
	}
	return nil
}

func validChild(child int, tree TreeSpec) bool {
	return child >= -len(tree.Leaves) && child < len(tree.Splits)
}

type treeWalk struct {
	tree   TreeSpec
	splits []bool
	leaves []bool
}

func (w *treeWalk) visit(child, depth int) error {
	if depth > maxDepth {
		return fmt.Errorf("tree depth exceeds %d", maxDepth)
	}
	if child < 0 {
		index := -child - 1
		if w.leaves[index] {
			return fmt.Errorf("repeated tree leaf")
		}
		w.leaves[index] = true
		return nil
	}
	if w.splits[child] {
		return fmt.Errorf("cycle or repeated tree split")
	}
	w.splits[child] = true
	split := w.tree.Splits[child]
	if err := w.visit(split.Left, depth+1); err != nil {
		return err
	}
	return w.visit(split.Right, depth+1)
}
