/*****************************************************************************/
/* a_star.go                                                                 */
/*****************************************************************************/
/*                           This file is part of:                           */
/*                                KAIJU ENGINE                               */
/*                          https://kaijuengine.org                          */
/*****************************************************************************/
/* MIT License                                                               */
/*                                                                           */
/* Copyright (c) 2023-present Kaiju Engine contributors (CONTRIBUTORS.md).   */
/* Copyright (c) 2015-2023 Brent Farris.                                     */
/*                                                                           */
/* May all those that this source may reach be blessed by the LORD and find  */
/* peace and joy in life.                                                    */
/* Everyone who drinks of this water will be thirsty again; but whoever      */
/* drinks of the water that I will give him shall never thirst; John 4:13-14 */
/*                                                                           */
/* Permission is hereby granted, free of charge, to any person obtaining a   */
/* copy of this software and associated documentation files (the "Software"),*/
/* to deal in the Software without restriction, including without limitation */
/* the rights to use, copy, modify, merge, publish, distribute, sublicense,  */
/* and/or sell copies of the Software, and to permit persons to whom the     */
/* Software is furnished to do so, subject to the following conditions:      */
/*                                                                           */
/* The above copyright, blessing, biblical verse, notice and                 */
/* this permission notice shall be included in all copies or                 */
/* substantial portions of the Software.                                     */
/*                                                                           */
/* THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS   */
/* OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF                */
/* MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.    */
/* IN NO EVENT SHALL THE /* AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY   */
/* CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT */
/* OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE     */
/* OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.                             */
/*****************************************************************************/

package navigation

import (
	"container/heap"
	"kaiju/matrix"
)

func AStar(grid Grid, start, end matrix.Vec3i) []*Node {
	if grid.IsBlocked(end) {
		end = findNearestUnblockedNode(grid, end)
		if end[matrix.Vx] == -1 && end[matrix.Vy] == -1 && end[matrix.Vz] == -1 {
			return nil
		}
	}
	openSet := make(PriorityQueue, 0)
	closedSet := make(map[[3]int32]bool)
	startNode := &Node{x: start[0], y: start[1], z: start[2]}
	endNode := &Node{x: end[0], y: end[1], z: end[2]}
	openSet = append(openSet, startNode)
	heap.Init(&openSet)
	for len(openSet) > 0 {
		current := heap.Pop(&openSet).(*Node)
		if current.x == endNode.x && current.y == endNode.y && current.z == endNode.z {
			path := make([]*Node, 0)
			for current != nil {
				path = append(path, current)
				current = current.parent
			}
			reversePath(path)
			return path
		}
		closedSet[[3]int32{current.x, current.y, current.z}] = true
		neighbors := getNeighbors(current, grid)
		for _, neighbor := range neighbors {
			if closedSet[[3]int32{neighbor.x, neighbor.y, neighbor.z}] {
				continue
			}
			tentativeG := current.g + 1.0 // Assuming uniform cost
			if !contains(openSet, neighbor) || tentativeG < neighbor.g {
				neighbor.g = tentativeG
				neighbor.h = heuristic(neighbor, endNode)
				neighbor.f = neighbor.g + neighbor.h
				neighbor.parent = current
				if !contains(openSet, neighbor) {
					heap.Push(&openSet, neighbor)
				}
			}
		}
	}
	return nil
}

func getNeighbors(node *Node, grid Grid) []*Node {
	neighbors := make([]*Node, 0)
	directions := [][3]int8{
		{1, 0, 0}, {-1, 0, 0}, {0, 1, 0}, {0, -1, 0}, {0, 0, 1}, {0, 0, -1},
		{1, 1, 0}, {-1, 1, 0}, {1, -1, 0}, {-1, -1, 0}, {1, 0, 1}, {-1, 0, 1},
		{1, 0, -1}, {-1, 0, -1}, {0, 1, 1}, {0, -1, 1}, {0, 1, -1}, {0, -1, -1},
		{1, 1, 1}, {-1, 1, 1}, {1, -1, 1}, {-1, -1, 1}, {1, 1, -1}, {-1, 1, -1},
		{1, -1, -1}, {-1, -1, -1},
	}
	for _, dir := range directions {
		x, y, z := node.x+int32(dir[0]), node.y+int32(dir[1]), node.z+int32(dir[2])
		if isValid(x, y, z, grid) {
			neighbors = append(neighbors, &Node{x: x, y: y, z: z})
		}
	}
	return neighbors
}

func findNearestUnblockedNode(grid Grid, blockedEnd [3]int32) matrix.Vec3i {
	visited := make(map[[3]int32]bool)
	queue := make([][3]int32, 0)
	queue = append(queue, blockedEnd)
	visited[blockedEnd] = true
	directions := [][3]int32{{1, 0, 0}, {-1, 0, 0}, {0, 1, 0}, {0, -1, 0}, {0, 0, 1}, {0, 0, -1}}
	for len(queue) > 0 {
		currentNode := queue[0]
		queue = queue[1:]
		for _, dir := range directions {
			x, y, z := currentNode[0]+dir[0], currentNode[1]+dir[1], currentNode[2]+dir[2]
			neighbor := [3]int32{x, y, z}
			if isValid(x, y, z, grid) && !visited[neighbor] {
				if grid[x][y][z] == 0 { // Found an unblocked node
					return neighbor
				}
				queue = append(queue, neighbor)
				visited[neighbor] = true
			}
		}
	}
	// Return a default value (e.g., (-1, -1, -1)) if no unblocked node is found
	return [3]int32{-1, -1, -1}
}

func isValid(x, y, z int32, grid Grid) bool {
	if x < 0 || x >= int32(len(grid)) || y < 0 || y >= int32(len(grid[0])) || z < 0 || z >= int32(len(grid[0][0])) {
		return false
	}
	return grid[x][y][z] == 0 // Assuming 0 represents an open cell
}

func heuristic(a, b *Node) float64 {
	// Euclidean distance as the heuristic
	dx := a.x - b.x
	dy := a.y - b.y
	dz := a.z - b.z
	return float64(dx*dx + dy*dy + dz*dz)
}

func reversePath(path []*Node) {
	for i, j := 0, len(path)-1; i < j; i, j = i+1, j-1 {
		path[i], path[j] = path[j], path[i]
	}
}

func contains(nodes PriorityQueue, node *Node) bool {
	for _, n := range nodes {
		if n.x == node.x && n.y == node.y && n.z == node.z {
			return true
		}
	}
	return false
}
