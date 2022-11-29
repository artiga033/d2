package d2sequence

import (
	"context"
	"testing"

	"oss.terrastruct.com/d2/d2graph"
	"oss.terrastruct.com/d2/lib/geo"
	"oss.terrastruct.com/d2/lib/log"
)

func TestBasicSequenceDiagram(t *testing.T) {
	// ┌────────┐              ┌────────┐
	// │   n1   │              │   n2   │
	// └────┬───┘              └────┬───┘
	//      │                       │
	//      ├───────────────────────►
	//      │                       │
	//      ◄───────────────────────┤
	//      │                       │
	//      ├───────────────────────►
	//      │                       │
	//      ◄───────────────────────┤
	//      │                       │
	g := d2graph.NewGraph(nil)
	n1 := g.Root.EnsureChild([]string{"n1"})
	n1.Box = geo.NewBox(nil, 100, 100)
	n2 := g.Root.EnsureChild([]string{"n2"})
	n2.Box = geo.NewBox(nil, 30, 30)

	g.Edges = []*d2graph.Edge{
		{
			Src: n1,
			Dst: n2,
		},
		{
			Src: n2,
			Dst: n1,
		},
		{
			Src: n1,
			Dst: n2,
		},
		{
			Src: n2,
			Dst: n1,
		},
	}
	nEdges := len(g.Edges)

	ctx := log.WithTB(context.Background(), t, nil)
	Layout(ctx, g)

	// asserts that actors were placed in the expected x order and at y=0
	actors := []*d2graph.Object{
		g.Objects[0],
		g.Objects[1],
	}
	for i := 1; i < len(actors); i++ {
		if actors[i].TopLeft.X < actors[i-1].TopLeft.X {
			t.Fatalf("expected actor[%d].TopLeft.X > actor[%d].TopLeft.X", i, i-1)
		}
		actorBottom := actors[i].TopLeft.Y + actors[i].Height
		prevActorBottom := actors[i-1].TopLeft.Y + actors[i-1].Height
		if actorBottom != prevActorBottom {
			t.Fatalf("expected actor[%d] and actor[%d] to be at the same bottom y", i, i-1)
		}
	}

	nExpectedEdges := nEdges + len(actors)
	if len(g.Edges) != nExpectedEdges {
		t.Fatalf("expected %d edges, got %d", nExpectedEdges, len(g.Edges))
	}

	// assert that edges were placed in y order and have the endpoints at their actors
	// uses `nEdges` because Layout creates some vertical edges to represent the actor lifeline
	for i := 0; i < nEdges; i++ {
		edge := g.Edges[i]
		if len(edge.Route) != 2 {
			t.Fatalf("expected edge[%d] to have only 2 points", i)
		}
		if edge.Route[0].Y != edge.Route[1].Y {
			t.Fatalf("expected edge[%d] to be a horizontal line", i)
		}
		if edge.Route[0].X != edge.Src.Center().X {
			t.Fatalf("expected edge[%d] source endpoint to be at the middle of the source actor", i)
		}
		if edge.Route[1].X != edge.Dst.Center().X {
			t.Fatalf("expected edge[%d] target endpoint to be at the middle of the target actor", i)
		}
		if i > 0 {
			prevEdge := g.Edges[i-1]
			if edge.Route[0].Y < prevEdge.Route[0].Y {
				t.Fatalf("expected edge[%d].TopLeft.Y > edge[%d].TopLeft.Y", i, i-1)
			}
		}
	}

	lastSequenceEdge := g.Edges[nEdges-1]
	for i := nEdges; i < nExpectedEdges; i++ {
		edge := g.Edges[i]
		if len(edge.Route) != 2 {
			t.Fatalf("expected edge[%d] to have only 2 points", i)
		}
		if edge.Route[0].X != edge.Route[1].X {
			t.Fatalf("expected edge[%d] to be a vertical line", i)
		}
		if edge.Route[0].X != edge.Src.Center().X {
			t.Fatalf("expected edge[%d] x to be at the actor center", i)
		}
		if edge.Route[0].Y != edge.Src.Height+edge.Src.TopLeft.Y {
			t.Fatalf("expected edge[%d] to start at the bottom of the source actor", i)
		}
		if edge.Route[1].Y < lastSequenceEdge.Route[0].Y {
			t.Fatalf("expected edge[%d] to end after the last sequence edge", i)
		}
	}
}

func TestActivationBoxesSequenceDiagram(t *testing.T) {
	//   ┌─────┐                 ┌─────┐
	//   │  a  │                 │  b  │
	//   └──┬──┘                 └──┬──┘
	//      ├┐────────────────────►┌┤
	//   t1 ││                     ││ t1
	//      ├┘◄────────────────────└┤
	//      ├┐──────────────────────►
	//   t2 ││                      │
	//      ├┘◄─────────────────────┤
	g := d2graph.NewGraph(nil)
	a := g.Root.EnsureChild([]string{"a"})
	a.Box = geo.NewBox(nil, 100, 100)
	a_t1 := a.EnsureChild([]string{"t1"})
	a_t2 := a.EnsureChild([]string{"t2"})
	b := g.Root.EnsureChild([]string{"b"})
	b.Box = geo.NewBox(nil, 30, 30)
	b_t1 := b.EnsureChild([]string{"t1"})

	g.Edges = []*d2graph.Edge{
		{
			Src: a_t1,
			Dst: b_t1,
		}, {
			Src: b_t1,
			Dst: a_t1,
		}, {
			Src: a_t2,
			Dst: b,
		}, {
			Src: b,
			Dst: a_t2,
		},
	}

	ctx := log.WithTB(context.Background(), t, nil)
	Layout(ctx, g)

	if a.Center().X != a_t1.Center().X {
		t.Fatal("expected a_t1.X = a.X")
	}
	if a.Center().X != a_t2.Center().X {
		t.Fatal("expected a_t2.X = a.X")
	}
	if b.Center().X != b_t1.Center().X {
		t.Fatal("expected b_t1.X = b.X")
	}
}
