[![GoDoc](https://godoc.org/github.com/cloudprivacylabs/opencypher?status.svg)](https://godoc.org/github.com/cloudprivacylabs/opencypher)
[![Go Report Card](https://goreportcard.com/badge/github.com/cloudprivacylabs/opencypher)](https://goreportcard.com/report/github.com/cloudprivacylabs/opencypher)
[![Build Status](https://github.com/cloudprivacylabs/opencypher/actions/workflows/CI.yml/badge.svg?branch=main)](https://github.com/cloudprivacylabs/opencypher/actions/workflows/CI.yml)
# Embedded openCypher interpreter and labeled property graphs

This Go module contains a openCypher interpreter partial
implementation and a labeled property graph implementation. The
labeled property graph package can be used independently from the
openCypher library.

This Go module is part of the [Layered Schema
Architecture](https://layeredschemas.org).

## Labeled Property Graph

This labeled property graph package implements the openCypher model of
labeled property graphs. A labeled property graph (LPG) contains nodes
and directed edges between those nodes. Every node contains:

  * Labels: Set of string tokens that usually identify the type of the
    node,
  * Properties: Key-value pairs.
  
Every edge contains:
  * A label: String token that identifies a relationship, and
  * Properties: Key-value pairs.
  
A graph indexes its nodes and edges, so finding a node, or a pattern
usually does not involve iterating through all possibilities. 

Create a graph using `NewOCGraph` function:

```
g := graph.NewOCGraph()
// Create two nodes
n1 := g.NewNode([]string{"label1"},map[string]interface{}{"prop": "value1" })
n2 := g.NewNode([]string{"label2"},map[string]interface{}{"prop": "value2" })
// Connect the two nodes with an edge
edge:=g.NewEdge(n1,n2,"relatedTo",nil)
```

The LPG library uses an iterator model to address nodes and edges
because the underlying algorithm to collect nodes and edges mathcing a
certain criteria may depend on the existence of indexes. Both incoming
and outgoing edges of nodes are accessible:

```
for edges:=n1.GetEdges(graph.OutgoingEdge); edges.Next(); {
  edge:=edges.Edge()
  // edge.GetTo() and edge.GetFrom() are the adjacent nodes
}
```


The graph indexes nodes by label, so access to nodes using labels is
fast. You can add additional indexes on properties:

```
g := graph.NewOCGraph()
// Index all nodes with property 'prop'
g.AddNodePropertyIndex("prop")

// This access should be fast
nodes := g.GetNodesWithProperty("prop")

// This will go through all nodes
slowNodes:= g.GetNodesWithProperty("propWithoutIndex")
```

Graph library supports searching patterns. The following example
searches for the pattern that match 

```
(:label1) -[]->({prop:value})`
```

and returns the head nodes for every matching path:

```
pattern := graph.Pattern{ 
 // Node containing label 'label1'
 {
   Labels: graph.NewStringSet("label1"),
 },
 // Edge of length 1
 {
   Min: 1, 
   Max: 1,
 },
 // Node with property prop=value
 {
   Properties: map[string]interface{} {"prop":"value"},
 }}
nodes, err:=pattern.FindNodes(g,nil)
```

## openCypher

At this point, this library provides partial support for openCypher
expressions. More support will be added as needed.

openCypher expressions are evaluated using an evaluation context. 

```
ectx:=NewEvalContext(grph)
val, err:=opencypher.ParseAndEvaluate(`match (n:label) return n`,ctx)
resultSet:=val.Value.(opencypher.ResultSet)
for _,row:=range resultSet.Rows {
  node:=row[n].Value.(graph.Node)
}
```

