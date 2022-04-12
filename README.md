### AVL - AVL tree
#### Jesse Kuang (jkuang at 21cn dot com)

[![GoDoc](https://godoc.org/github.com/kjx98/go-avl.git?status.svg)](https://godoc.org/github.com/kjx98/go-avl.git)

A generic type Go AVL tree implementation, forked from [Yawning Angel][1],
derived from
[Eric Biggers' C code][2], in the spirt of [the runtime library's containers][3].

Features:

 * Size
 * Insertion
 * Deletion
 * Search
 * In-order traversal (forward and backward) with an iterator or callback.
 * Non-recursive.

Generic Improve:
 * Generic type improved over 20% compared to interface{} generic func
 * Benchmark cmp
<pre>
benchmark                  old ns/op     new ns/op     delta
BenchmarkAVLInsert         1003          674           -32.83%
BenchmarkAVLFind           880           687           -21.92%
BenchmarkAVLDeleteLeft     0.16          69.0          +42812.26%
</pre>

Note:

 * The package itself is free from external dependencies, the unit tests use
   [testify][4]. dependencies removed

[1]: https://github.com/Yawning/avl
[2]: https://github.com/ebiggers/avl_tree
[3]: https://golang.org/pkg/container
[4]: https://github.com/stretchr/testify
