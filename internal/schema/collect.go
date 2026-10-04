package schema

// CollectTypes returns every distinct named object type reachable from
// root, in dependency order: a type always appears after the nested types
// it references, and the root type (or the element type, if root is an
// array) is always last. Each distinct *Node appears exactly once.
func CollectTypes(root *Node) []*Node {
	seen := make(map[*Node]bool)
	var order []*Node

	var visit func(n *Node)
	visit = func(n *Node) {
		if n == nil {
			return
		}
		switch n.Kind {
		case KindObject:
			if seen[n] {
				return
			}
			seen[n] = true
			for _, f := range n.Fields {
				visit(f.Node)
			}
			order = append(order, n)
		case KindArray:
			visit(n.Element)
		}
	}

	visit(root)
	return order
}
