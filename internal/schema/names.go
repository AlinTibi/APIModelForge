package schema

// assignNames walks the inferred tree and gives every object node a
// unique, PascalCase type Name. hint is the best candidate name for the
// node at the current position (the user-supplied root name, or the
// PascalCase/singularized field name that contains it). used tracks every
// name already handed out, so colliding names for genuinely different
// shapes get a numeric suffix instead of silently reusing a type.
func assignNames(n *Node, hint string, used map[string]*Node) {
	if n == nil {
		return
	}

	switch n.Kind {
	case KindObject:
		n.Name = allocateName(used, PascalCase(hint), n)
		for _, f := range n.Fields {
			assignNames(f.Node, PascalCase(f.JSONName), used)
		}
	case KindArray:
		assignNames(n.Element, singularize(hint), used)
	}
}

func allocateName(used map[string]*Node, base string, n *Node) string {
	if base == "" {
		base = "Model"
	}

	if owner, ok := used[base]; !ok || owner == n {
		used[base] = n
		return base
	}

	for i := 2; ; i++ {
		candidate := base + itoa(i)
		if owner, ok := used[candidate]; !ok || owner == n {
			used[candidate] = n
			return candidate
		}
	}
}

// itoa avoids pulling in strconv just for small positive integers.
func itoa(i int) string {
	if i == 0 {
		return "0"
	}
	var digits []byte
	for i > 0 {
		digits = append([]byte{byte('0' + i%10)}, digits...)
		i /= 10
	}
	return string(digits)
}
