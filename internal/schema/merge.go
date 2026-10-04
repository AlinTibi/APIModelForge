package schema

// Merge combines two Node observations for the same logical position
// (the same object field across multiple array elements, or successive
// elements of the same array) into one Node that safely describes both.
//
// Rules:
//   - a nil accumulator is replaced by the other node outright
//   - null + T => T, marked Nullable
//   - matching kinds merge structurally (objects merge fields, arrays
//     merge element types, scalars just combine their Nullable flag)
//   - int + float => float (numeric widening)
//   - any other kind conflict => KindAny, the universal fallback
func Merge(a, b *Node) *Node {
	if a == nil {
		return b
	}
	if b == nil {
		return a
	}

	if a.Kind == KindNull && b.Kind == KindNull {
		return &Node{Kind: KindNull, Nullable: true}
	}
	if a.Kind == KindNull {
		clone := *b
		clone.Nullable = true
		return &clone
	}
	if b.Kind == KindNull {
		clone := *a
		clone.Nullable = true
		return &clone
	}

	nullable := a.Nullable || b.Nullable

	if a.Kind == b.Kind {
		switch a.Kind {
		case KindObject:
			merged := mergeObjects(a, b)
			merged.Nullable = nullable
			return merged
		case KindArray:
			return &Node{Kind: KindArray, Element: Merge(a.Element, b.Element), Nullable: nullable}
		default:
			return &Node{Kind: a.Kind, Nullable: nullable}
		}
	}

	if isNumeric(a.Kind) && isNumeric(b.Kind) {
		return &Node{Kind: KindFloat, Nullable: nullable}
	}

	// Any other kind mismatch (object vs string, array vs object, bool
	// vs number, ...) has no safe common representation.
	return &Node{Kind: KindAny, Nullable: nullable}
}

func isNumeric(k Kind) bool {
	return k == KindInt || k == KindFloat
}

func mergeObjects(a, b *Node) *Node {
	result := &Node{Kind: KindObject}

	bByName := make(map[string]*Node, len(b.Fields))
	for _, f := range b.Fields {
		bByName[f.JSONName] = f.Node
	}
	seenInA := make(map[string]bool, len(a.Fields))

	for _, af := range a.Fields {
		seenInA[af.JSONName] = true
		if bf, ok := bByName[af.JSONName]; ok {
			result.Fields = append(result.Fields, &Field{JSONName: af.JSONName, Node: Merge(af.Node, bf)})
		} else {
			result.Fields = append(result.Fields, &Field{JSONName: af.JSONName, Node: withNullable(af.Node)})
		}
	}

	for _, bf := range b.Fields {
		if seenInA[bf.JSONName] {
			continue
		}
		result.Fields = append(result.Fields, &Field{JSONName: bf.JSONName, Node: withNullable(bf.Node)})
	}

	return result
}

// withNullable clones a node as Nullable, used when a field is present in
// one sampled object but absent from another: the generated property must
// tolerate a missing value.
func withNullable(n *Node) *Node {
	clone := *n
	clone.Nullable = true
	return &clone
}
