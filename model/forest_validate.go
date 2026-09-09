package model

import "fmt"

const maxForestSamples = 100000

func validateForest(p ForestParameters) error {
	if p.Features < 1 || p.Features > MaxFeatures || len(p.Trees) < 1 || len(p.Trees) > MaxForestTrees {
		return fmt.Errorf("forest requires 1..%d features and 1..%d trees", MaxFeatures, MaxForestTrees)
	}
	nodes := 0
	for i, tree := range p.Trees {
		if len(tree) == 0 || len(tree) > MaxForestNodes-nodes {
			return fmt.Errorf("forest has an empty tree or exceeds the node limit")
		}
		nodes += len(tree)
		end, err := validateSubtree(tree, 0, 0, p.Features)
		if err != nil {
			return fmt.Errorf("forest tree %d: %w", i, err)
		}
		if end != len(tree) {
			return fmt.Errorf("forest tree %d has unreachable nodes", i)
		}
	}
	return nil
}

func validateSubtree(tree []ForestNode, index, depth, width int) (int, error) {
	if index < 0 || index >= len(tree) || depth > MaxForestDepth {
		return 0, fmt.Errorf("invalid node index or tree depth")
	}
	node := tree[index]
	if err := validateNodeFields(node); err != nil {
		return 0, err
	}
	if node.Feature == -1 {
		if !validLeaf(node) {
			return 0, fmt.Errorf("leaf must have no split or children")
		}
		return index + 1, nil
	}
	if node.Feature < 0 || node.Feature >= width || node.Left != index+1 {
		return 0, fmt.Errorf("invalid split feature or preorder left child")
	}
	return validateChildren(tree, node, depth, width)
}

func validateNodeFields(node ForestNode) error {
	if node.Samples < 1 || node.Samples > maxForestSamples || node.Positive < 0 || node.Positive > node.Samples ||
		!finite(node.Threshold) {
		return fmt.Errorf("invalid node counts or threshold")
	}
	return nil
}

func validLeaf(node ForestNode) bool {
	return node.Left == -1 && node.Right == -1 && node.Threshold == 0
}

func validateChildren(tree []ForestNode, node ForestNode, depth, width int) (int, error) {
	end, err := validateSubtree(tree, node.Left, depth+1, width)
	if err != nil {
		return 0, err
	}
	if node.Right != end {
		return 0, fmt.Errorf("right child must follow the complete left subtree")
	}
	end, err = validateSubtree(tree, node.Right, depth+1, width)
	if err != nil {
		return 0, err
	}
	left, right := tree[node.Left], tree[node.Right]
	if node.Samples != left.Samples+right.Samples || node.Positive != left.Positive+right.Positive {
		return 0, fmt.Errorf("parent counts must equal child totals")
	}
	return end, nil
}
