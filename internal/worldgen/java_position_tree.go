package worldgen

// HashMap tree bins keep a linked iteration list in addition to a red-black
// lookup tree. Treeification moves the root to the front; later insertion
// links the new entry immediately after its lookup parent.
type javaPositionTree struct {
	p                   FeaturePosition
	red                 bool
	parent, left, right *javaPositionTree
}

func positionTreeLess(a, b FeaturePosition) bool {
	ah, bh := int32(positionHash(a)), int32(positionHash(b))
	if ah != bh {
		return ah < bh
	}
	if a.Y != b.Y {
		return a.Y < b.Y
	}
	if a.Z != b.Z {
		return a.Z < b.Z
	}
	return a.X < b.X
}

func rotatePositionLeft(root, p *javaPositionTree) *javaPositionTree {
	r := p.right
	p.right = r.left
	if r.left != nil {
		r.left.parent = p
	}
	r.parent = p.parent
	if p.parent == nil {
		root = r
		r.red = false
	} else if p.parent.left == p {
		p.parent.left = r
	} else {
		p.parent.right = r
	}
	r.left = p
	p.parent = r
	return root
}

func rotatePositionRight(root, p *javaPositionTree) *javaPositionTree {
	l := p.left
	p.left = l.right
	if l.right != nil {
		l.right.parent = p
	}
	l.parent = p.parent
	if p.parent == nil {
		root = l
		l.red = false
	} else if p.parent.right == p {
		p.parent.right = l
	} else {
		p.parent.left = l
	}
	l.right = p
	p.parent = l
	return root
}

func balancePositionInsertion(root, x *javaPositionTree) *javaPositionTree {
	x.red = true
	for {
		p := x.parent
		if p == nil {
			x.red = false
			return x
		}
		g := p.parent
		if !p.red || g == nil {
			return root
		}
		if p == g.left {
			u := g.right
			if u != nil && u.red {
				u.red = false
				p.red = false
				g.red = true
				x = g
				continue
			}
			if x == p.right {
				root = rotatePositionLeft(root, p)
				x = p
				p = x.parent
				g = p.parent
			}
			p.red = false
			if g != nil {
				g.red = true
				root = rotatePositionRight(root, g)
			}
		} else {
			u := g.left
			if u != nil && u.red {
				u.red = false
				p.red = false
				g.red = true
				x = g
				continue
			}
			if x == p.left {
				root = rotatePositionRight(root, p)
				x = p
				p = x.parent
				g = p.parent
			}
			p.red = false
			if g != nil {
				g.red = true
				root = rotatePositionLeft(root, g)
			}
		}
	}
}

func (set *javaPositionSet) moveTreeRoot(i int) {
	p := set.trees[i].p
	list := set.buckets[i]
	for j, q := range list {
		if q == p {
			copy(list[1:j+1], list[:j])
			list[0] = p
			return
		}
	}
}

func (set *javaPositionSet) treeify(i int) {
	var root *javaPositionTree
	for _, p := range set.buckets[i] {
		x := &javaPositionTree{p: p}
		if root == nil {
			root = x
			continue
		}
		parent := root
		for {
			left := positionTreeLess(p, parent.p)
			child := parent.right
			if left {
				child = parent.left
			}
			if child != nil {
				parent = child
				continue
			}
			x.parent = parent
			if left {
				parent.left = x
			} else {
				parent.right = x
			}
			root = balancePositionInsertion(root, x)
			break
		}
	}
	set.trees[i] = root
	if root != nil {
		set.moveTreeRoot(i)
	}
}

func (set *javaPositionSet) addTree(i int, p FeaturePosition) {
	root := set.trees[i]
	parent := root
	for {
		left := positionTreeLess(p, parent.p)
		child := parent.right
		if left {
			child = parent.left
		}
		if child != nil {
			parent = child
			continue
		}
		x := &javaPositionTree{p: p, parent: parent}
		if left {
			parent.left = x
		} else {
			parent.right = x
		}
		list := set.buckets[i]
		for j, q := range list {
			if q == parent.p {
				list = append(list, FeaturePosition{})
				copy(list[j+2:], list[j+1:])
				list[j+1] = p
				break
			}
		}
		set.buckets[i] = list
		set.trees[i] = balancePositionInsertion(root, x)
		set.moveTreeRoot(i)
		return
	}
}
