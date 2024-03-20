package collision

import "kaiju/matrix"

type HitObject interface {
	Bounds() AABB
	RayIntersect(ray Ray) bool
	AABBIntersect(bounds AABB) bool
}

type OctreeNode struct {
	Center    matrix.Vec3
	HalfWidth float64
	Children  [8]*OctreeNode
	Objects   []HitObject
}

type Octree struct {
	Root *OctreeNode
}

func NewOctreeNode(center matrix.Vec3, halfWidth float64, maxDepth int) *OctreeNode {
	if maxDepth == 0 {
		return nil
	}
	n := &OctreeNode{
		Center:    center,
		HalfWidth: halfWidth,
		Objects:   make([]HitObject, 0),
	}
	offset := matrix.Vec3{}
	step := float32(halfWidth * 0.5)
	for i := 0; i < 8; i++ {
		offset = matrix.Vec3{step, step, step}
		if i&1 == 1 {
			offset[matrix.Vx] *= -1
		}
		if i&2 == 2 {
			offset[matrix.Vy] *= -1
		}
		if i&4 == 4 {
			offset[matrix.Vz] *= -1
		}
		n.Children[i] = NewOctreeNode(center.Add(offset), halfWidth*0.5, maxDepth-1)
	}
	return n
}

func NewOctree(center matrix.Vec3, halfWidth float64, stopDepth int) Octree {
	return Octree{
		Root: NewOctreeNode(center, halfWidth, stopDepth),
	}
}

func (node *OctreeNode) Insert(obj HitObject) {
	index := 0
	straddle := 0
	for i := 0; i < 3; i++ {
		delta := obj.Bounds().Center[i] - node.Center[i]
		if matrix.Abs(delta) <= obj.Bounds().Extent[i] {
			straddle = 1
			break
		}
		if delta > 0 {
			index |= 1 << uint(i)
		}
	}
	if straddle == 0 && node.Children[index] != nil {
		node.Children[index].Insert(obj)
	} else {
		node.Objects = append(node.Objects, obj)
	}
}

func (tree *Octree) Query(bounds AABB) []HitObject {
	return tree.queryRecursive(tree.Root, bounds)
}

func (tree *Octree) RayIntersect(ray Ray) []HitObject {
	return tree.rayIntersectRecursive(tree.Root, ray)
}

func (tree *Octree) insertRecursive(node *OctreeNode, obj HitObject, depth int) {
	node.Objects = append(node.Objects, obj)
	for i := 0; i < 8; i++ {
		childBounds := node.calculateChildBounds(i)
		if childBounds.AABBIntersect(obj.Bounds()) {
			if node.Children[i] == nil {
				node.Children[i] = NewOctreeNode(childBounds)
			}
			tree.insertRecursive(node.Children[i], obj, depth+1)
		}
	}
}

func (tree *Octree) queryRecursive(node *OctreeNode, bounds AABB) []HitObject {
	result := make([]HitObject, 0)
	if node.Bounds.AABBIntersect(bounds) {
		return result
	}
	for _, obj := range node.Objects {
		if obj.AABBIntersect(bounds) {
			result = append(result, obj)
		}
	}
	for _, child := range node.Children {
		if child != nil {
			result = append(result, tree.queryRecursive(child, bounds)...)
		}
	}
	return result
}

func (node *OctreeNode) calculateChildBounds(index int) AABB {
	bMin := node.Bounds.Min()
	bMax := node.Bounds.Max()
	mid := matrix.Vec3{
		matrix.Vx: (bMin.X() + bMax.X()) / 2,
		matrix.Vy: (bMin.Y() + bMax.Y()) / 2,
		matrix.Vz: (bMin.Z() + bMax.Z()) / 2,
	}
	var childBounds AABB
	switch index {
	case 0:
		childBounds = AABBFromMinMax(bMin, mid)
	case 1:
		childBounds = AABBFromMinMax(matrix.Vec3{mid.X(), bMin.Y(), bMin.Z()},
			matrix.Vec3{bMax.X(), mid.Y(), mid.Z()})
	case 2:
		childBounds = AABBFromMinMax(matrix.Vec3{bMin.X(), mid.Y(), bMin.Z()},
			matrix.Vec3{mid.X(), bMax.Y(), mid.Z()})
	case 3:
		childBounds = AABBFromMinMax(matrix.Vec3{mid.X(), mid.Y(), bMin.Z()},
			matrix.Vec3{bMax.X(), bMax.Y(), mid.Z()})
	case 4:
		childBounds = AABBFromMinMax(matrix.Vec3{bMin.X(), bMin.Y(), mid.Z()},
			matrix.Vec3{mid.X(), mid.Y(), bMax.Z()})
	case 5:
		childBounds = AABBFromMinMax(matrix.Vec3{mid.X(), bMin.Y(), mid.Z()},
			matrix.Vec3{bMax.X(), mid.Y(), bMax.Z()})
	case 6:
		childBounds = AABBFromMinMax(matrix.Vec3{bMin.X(), mid.Y(), mid.Z()},
			matrix.Vec3{mid.X(), bMax.Y(), bMax.Z()})
	case 7:
		childBounds = AABBFromMinMax(mid, bMax)
	}
	return childBounds
}

func (tree *Octree) rayIntersectRecursive(node *OctreeNode, ray Ray) []HitObject {
	result := make([]HitObject, 0)
	if _, ok := node.Bounds.RayHit(ray); !ok {
		return result
	}
	for _, obj := range node.Objects {
		if obj.RayIntersect(ray) {
			result = append(result, obj)
		}
	}
	for _, child := range node.Children {
		if child != nil {
			result = append(result, tree.rayIntersectRecursive(child, ray)...)
		}
	}
	return result
}
