package constraineddelaunay

import (
	"encoding/hex"
	"log"
	"reflect"
	"strconv"
	"testing"

	"github.com/go-spatial/geom"
	"github.com/go-spatial/geom/encoding/wkb"
	"github.com/go-spatial/geom/encoding/wkt"
	"github.com/go-spatial/geom/planar/triangulate"
	"github.com/go-spatial/geom/planar/triangulate/quadedge"
)

func TestFindIntersectingTriangle(t *testing.T) {
	type tcase struct {
		// provided for readability
		inputWKT string
		// this can be removed if/when geom has a WKT decoder.
		// A simple website for performing conversions:
		// https://rodic.fr/blog/online-conversion-between-geometric-formats/
		inputWKB         string
		searchFrom       geom.Line
		expectedTriangle string
		err error
	}

	fn := func(t *testing.T, tc tcase) {
		bytes, err := hex.DecodeString(tc.inputWKB)
		if err != nil {
			t.Fatalf("error decoding hex string: %v", err)
			return
		}
		g, err := wkb.DecodeBytes(bytes)
		if err != nil {
			t.Fatalf("error decoding WKB: %v", err)
			return
		}

		uut := new(Triangulator)
		uut.tolerance = 1e-6
		// perform self consistency validation while building the
		// triangulation.
		uut.validate = true
		uut.insertSites(g)

		// find the triangle
		tri, err := uut.findIntersectingTriangle(triangulate.NewSegment(tc.searchFrom))
		if reflect.TypeOf(err) != reflect.TypeOf(tc.err) {
			t.Fatalf("error, expected %v got %v", reflect.TypeOf(tc.err), reflect.TypeOf(err))
			return
		}

		if tc.err == nil {
			qeStr := tri.String()
			if qeStr != tc.expectedTriangle {
				t.Fatalf("error, expected %v got %v", tc.expectedTriangle, qeStr)
			}
		}
	}
	testcases := []tcase{
		{
			inputWKT:         `MULTIPOINT (10 10, 10 20, 20 20, 20 10, 20 0, 10 0, 0 0, 0 10, 0 20)`,
			inputWKB:         `010400000009000000010100000000000000000024400000000000002440010100000000000000000024400000000000003440010100000000000000000034400000000000003440010100000000000000000034400000000000002440010100000000000000000034400000000000000000010100000000000000000024400000000000000000010100000000000000000000000000000000000000010100000000000000000000000000000000002440010100000000000000000000000000000000003440`,
			searchFrom:       geom.Line{{0, 0}, {10, 10}},
			expectedTriangle: `[[0 0],[0 10],[10 0]]`,
		},
		{
			inputWKT:         `MULTIPOINT (10 10, 10 20, 20 20, 20 10, 20 0, 10 0, 0 0, 0 10, 0 20)`,
			inputWKB:         `010400000009000000010100000000000000000024400000000000002440010100000000000000000024400000000000003440010100000000000000000034400000000000003440010100000000000000000034400000000000002440010100000000000000000034400000000000000000010100000000000000000024400000000000000000010100000000000000000000000000000000000000010100000000000000000000000000000000002440010100000000000000000000000000000000003440`,
			searchFrom:       geom.Line{{10, 0}, {0, 20}},
			expectedTriangle: `[[10 0],[0 10],[10 10]]`,
		},
		{
			inputWKT:         `MULTIPOINT (10 10, 10 20, 20 20, 20 10, 20 0, 10 0, 0 0, 0 10, 0 20)`,
			inputWKB:         `010400000009000000010100000000000000000024400000000000002440010100000000000000000024400000000000003440010100000000000000000034400000000000003440010100000000000000000034400000000000002440010100000000000000000034400000000000000000010100000000000000000024400000000000000000010100000000000000000000000000000000000000010100000000000000000000000000000000002440010100000000000000000000000000000000003440`,
			searchFrom:       geom.Line{{10, 10}, {0, 0}},
			expectedTriangle: `[[10 10],[10 0],[0 10]]`,
		},
		{
			inputWKT:         `MULTIPOINT (10 10, 10 20, 20 20, 20 10, 20 0, 10 0, 0 0, 0 10, 0 20)`,
			inputWKB:         `010400000009000000010100000000000000000024400000000000002440010100000000000000000024400000000000003440010100000000000000000034400000000000003440010100000000000000000034400000000000002440010100000000000000000034400000000000000000010100000000000000000024400000000000000000010100000000000000000000000000000000000000010100000000000000000000000000000000002440010100000000000000000000000000000000003440`,
			searchFrom:       geom.Line{{10, 10}, {10, 20}},
			expectedTriangle: `[[10 10],[10 20],[20 10]]`,
		},
		{
			inputWKT:         `MULTIPOINT (10 10, 10 20, 20 20, 20 10, 20 0, 10 0, 0 0, 0 10, 0 20)`,
			inputWKB:         `010400000009000000010100000000000000000024400000000000002440010100000000000000000024400000000000003440010100000000000000000034400000000000003440010100000000000000000034400000000000002440010100000000000000000034400000000000000000010100000000000000000024400000000000000000010100000000000000000000000000000000000000010100000000000000000000000000000000002440010100000000000000000000000000000000003440`,
			searchFrom:       geom.Line{{1000, 1000}, {10000, 20000}},
			expectedTriangle: ``,
			err: quadedge.ErrLocateFailure{},
		},
	}

	for i, tc := range testcases {
		tc := tc
		t.Run(strconv.FormatInt(int64(i), 10), func(t *testing.T) { fn(t, tc) })
	}
}

func TestDeleteEdge(t *testing.T) {
	type tcase struct {
		// provided for readability
		inputWKT string
		// this can be removed if/when geom has a WKT decoder.
		// A simple website for performing conversions:
		// https://rodic.fr/blog/online-conversion-between-geometric-formats/
		inputWKB string
		deleteMe geom.Line
	}

	fn := func(t *testing.T, tc tcase) {
		bytes, err := hex.DecodeString(tc.inputWKB)
		if err != nil {
			t.Fatalf("error decoding hex string, expected nil got %v", err)
			return
		}
		g, err := wkb.DecodeBytes(bytes)
		if err != nil {
			t.Fatalf("error decoding WKB, expected nil got %v", err)
			return
		}

		uut := new(Triangulator)
		uut.tolerance = 1e-6
		// perform self consistency validation while building the
		// triangulation.
		uut.validate = true
		uut.InsertGeometry(g)
		e, err := uut.LocateSegment(quadedge.Vertex(tc.deleteMe[0]), quadedge.Vertex(tc.deleteMe[1]))
		if err != nil {
			t.Fatalf("error locating segment, expected nil got %v", err)
			return
		}

		err = uut.Validate()
		if err != nil {
			t.Errorf("error validating triangulation, expected nil got %v", err)
			return
		}

		if err = uut.deleteEdge(e); err != nil {
			t.Errorf("error deleting edge, expected nil got %v", err)
		}
		// we know validateTriangles will fail, so just validate the indexes.
		err = uut.validateVertexIndex()
		if err != nil {
			t.Errorf("error validating triangulation after delete, expected nil got %v", err)
			return
		}

		// this edge shouldn't exist anymore.
		_, err = uut.LocateSegment(quadedge.Vertex(tc.deleteMe[0]), quadedge.Vertex(tc.deleteMe[1]))
		if err == nil {
			t.Fatalf("error locating segment, expected %v got nil", quadedge.ErrLocateFailure{})
			return
		}

	}
	testcases := []tcase{
		{
			inputWKT: `MULTIPOINT (10 10, 10 20, 20 20, 20 10, 20 0, 10 0, 0 0, 0 10, 0 20)`,
			inputWKB: `010400000009000000010100000000000000000024400000000000002440010100000000000000000024400000000000003440010100000000000000000034400000000000003440010100000000000000000034400000000000002440010100000000000000000034400000000000000000010100000000000000000024400000000000000000010100000000000000000000000000000000000000010100000000000000000000000000000000002440010100000000000000000000000000000000003440`,
			deleteMe: geom.Line{{0, 10}, {10, 0}},
		},
		{
			inputWKT: `MULTIPOINT (10 10, 10 20, 20 20, 20 10, 20 0, 10 0, 0 0, 0 10, 0 20)`,
			inputWKB: `010400000009000000010100000000000000000024400000000000002440010100000000000000000024400000000000003440010100000000000000000034400000000000003440010100000000000000000034400000000000002440010100000000000000000034400000000000000000010100000000000000000024400000000000000000010100000000000000000000000000000000000000010100000000000000000000000000000000002440010100000000000000000000000000000000003440`,
			deleteMe: geom.Line{{0, 10}, {10, 0}},
		},
	}

	for i, tc := range testcases {
		tc := tc
		t.Run(strconv.FormatInt(int64(i), 10), func(t *testing.T) { fn(t, tc) })
	}
}

func TestIntersection(t *testing.T) {
	type tcase struct {
		l1            triangulate.Segment
		l2            triangulate.Segment
		intersection  quadedge.Vertex
		expectedError error
	}

	fn := func(t *testing.T, tc tcase) {
		uut := new(Triangulator)
		uut.tolerance = 1e-2
		v, err := uut.intersection(tc.l1, tc.l2)
		if err != tc.expectedError {
			t.Errorf("error intersecting line segments, expected %v got %v", tc.expectedError, err)
			return
		}

		if v.Equals(tc.intersection) == false {
			t.Errorf("error validating intersection, expected %v got %v", tc.intersection, v)
		}
	}
	testcases := []tcase{
		{
			l1:            triangulate.NewSegment(geom.Line{{0, 1}, {2, 3}}),
			l2:            triangulate.NewSegment(geom.Line{{1, 1}, {0, 2}}),
			intersection:  quadedge.Vertex{0.5, 1.5},
			expectedError: nil,
		},
		{
			l1:            triangulate.NewSegment(geom.Line{{0, 1}, {2, 4}}),
			l2:            triangulate.NewSegment(geom.Line{{1, 1}, {0, 2}}),
			intersection:  quadedge.Vertex{0.4, 1.6},
			expectedError: nil,
		},
		{
			l1:            triangulate.NewSegment(geom.Line{{0, 1}, {2, 3}}),
			l2:            triangulate.NewSegment(geom.Line{{1, 1}, {2, 2}}),
			intersection:  quadedge.Vertex{0, 0},
			expectedError: ErrLinesDoNotIntersect,
		},
		{
			l1:            triangulate.NewSegment(geom.Line{{3, 5}, {3, 6}}),
			l2:            triangulate.NewSegment(geom.Line{{1, 4.995}, {4, 4.995}}),
			intersection:  quadedge.Vertex{3, 5},
			expectedError: nil,
		},
	}

	for i, tc := range testcases {
		tc := tc
		t.Run(strconv.FormatInt(int64(i), 10), func(t *testing.T) { fn(t, tc) })
	}
}

/*
TestTriangulation test cases test for small constrained triangulations and
edge cases
*/
func TestTriangulation(t *testing.T) {
	type tcase struct {
		// provided for readability
		inputWKT string
		// this can be removed if/when geom has a WKT decoder.
		// A simple website for performing conversions:
		// https://rodic.fr/blog/online-conversion-between-geometric-formats/
		inputWKB      string
		expectedEdges string
		expectedTris  string
	}

	// to change the flags on the default logger
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	fn := func(t *testing.T, tc tcase) {
		bytes, err := hex.DecodeString(tc.inputWKB)
		if err != nil {
			t.Fatalf("error decoding hex string: %v", err)
			return
		}
		g, err := wkb.DecodeBytes(bytes)
		if err != nil {
			t.Fatalf("error decoding WKB: %v", err)
			return
		}

		uut := new(Triangulator)
		uut.tolerance = 1e-6
		uut.validate = true
		err = uut.InsertGeometry(g)
		if err != nil {
			t.Fatalf("error inserting segments, expected nil got %v", err)
		}

		edges := uut.GetEdges()
		edgesWKT, err := wkt.Encode(edges)
		if err != nil {
			t.Errorf("error, expected nil got %v", err)
			return
		}
		if edgesWKT != tc.expectedEdges {
			t.Errorf("error, expected %v got %v", tc.expectedEdges, edgesWKT)
			return
		}

		tris, err := uut.GetTriangles()
		if err != nil {
			t.Errorf("error, expected nil got %v", err)
			return
		}
		trisWKT, err := wkt.Encode(tris)
		if err != nil {
			t.Errorf("error, expected nil got %v", err)
			return
		}
		if trisWKT != tc.expectedTris {
			t.Errorf("error, expected %v got %v", tc.expectedTris, trisWKT)
			return
		}
	}
	testcases := []tcase{
		{
			// should create a triangulation w/ a vertical line (2 5, 2 -5).
			// The unconstrained version has a horizontal line
			inputWKT:      `LINESTRING(0 0, 2 5, 2 -5, 5 0)`,
			inputWKB:      `0102000000040000000000000000000000000000000000000000000000000000400000000000001440000000000000004000000000000014c000000000000014400000000000000000`,
			expectedEdges: `MULTILINESTRING ((2 5,5 0),(0 0,2 5),(0 0,2 -5),(2 -5,5 0),(2 -5,2 5))`,
			expectedTris:  `MULTIPOLYGON (((0 0,2 -5,2 5,0 0)),((2 5,2 -5,5 0,2 5)))`,
		},
		{
			// a horizontal rectangle w/ one diagonal line. The diagonal line
			// should be maintained and the top/bottom re-triangulated.
			inputWKT:      `MULTILINESTRING((0 0,0 1,1 1.1,2 1,2 0,1 0.1,0 0),(0 1,2 0))`,
			inputWKB:      `010500000002000000010200000007000000000000000000000000000000000000000000000000000000000000000000f03f000000000000f03f9a9999999999f13f0000000000000040000000000000f03f00000000000000400000000000000000000000000000f03f9a9999999999b93f000000000000000000000000000000000102000000020000000000000000000000000000000000f03f00000000000000400000000000000000`,
			expectedEdges: `MULTILINESTRING ((1 1.1,2 1),(0 1,1 1.1),(0 0,0 1),(0 0,2 0),(2 0,2 1),(1 1.1,2 0),(0 1,2 0),(1 0.1,2 0),(0 1,1 0.1),(0 0,1 0.1))`,
			expectedTris:  `MULTIPOLYGON (((0 1,0 0,1 0.1,0 1)),((0 1,1 0.1,2 0,0 1)),((0 1,2 0,1 1.1,0 1)),((1 1.1,2 0,2 1,1 1.1)),((0 0,2 0,1 0.1,0 0)))`,
		},
		{
			// an egg shape with one horizontal line. The horizontal line
			// should be maintained and the top/bottom re-triangulated.
			inputWKT:      `MULTILINESTRING((0 0,-0.1 0.5,0 1,0.5 1.2,1 1.3,1.5 1.2,2 1,2.1 0.5,2 0,1.5 -0.2,1 -0.3,0.5 -0.2,0 0),(-0.1 0.5,2.1 0.5))`,
			inputWKB:      `01050000000200000001020000000d000000000000000000000000000000000000009a9999999999b9bf000000000000e03f0000000000000000000000000000f03f000000000000e03f333333333333f33f000000000000f03fcdccccccccccf43f000000000000f83f333333333333f33f0000000000000040000000000000f03fcdcccccccccc0040000000000000e03f00000000000000400000000000000000000000000000f83f9a9999999999c9bf000000000000f03f333333333333d3bf000000000000e03f9a9999999999c9bf000000000000000000000000000000000102000000020000009a9999999999b9bf000000000000e03fcdcccccccccc0040000000000000e03f`,
			expectedEdges: `MULTILINESTRING ((1.5 1.2,2 1),(1 1.3,1.5 1.2),(0.5 1.2,1 1.3),(0 1,0.5 1.2),(-0.1 0.5,0 1),(-0.1 0.5,0 0),(0 0,0.5 -0.2),(0.5 -0.2,1 -0.3),(1 -0.3,1.5 -0.2),(1.5 -0.2,2 0),(2 0,2.1 0.5),(2 1,2.1 0.5),(1.5 1.2,2.1 0.5),(1 1.3,2.1 0.5),(-0.1 0.5,2.1 0.5),(-0.1 0.5,1 1.3),(-0.1 0.5,0.5 1.2),(1 -0.3,2.1 0.5),(-0.1 0.5,1 -0.3),(1.5 -0.2,2.1 0.5),(-0.1 0.5,0.5 -0.2))`,
			expectedTris:  `MULTIPOLYGON (((0 1,-0.1 0.5,0.5 1.2,0 1)),((0.5 1.2,-0.1 0.5,1 1.3,0.5 1.2)),((1 1.3,-0.1 0.5,2.1 0.5,1 1.3)),((1 1.3,2.1 0.5,1.5 1.2,1 1.3)),((1.5 1.2,2.1 0.5,2 1,1.5 1.2)),((1 -0.3,1.5 -0.2,2.1 0.5,1 -0.3)),((1 -0.3,2.1 0.5,-0.1 0.5,1 -0.3)),((1 -0.3,-0.1 0.5,0.5 -0.2,1 -0.3)),((0.5 -0.2,-0.1 0.5,0 0,0.5 -0.2)),((2.1 0.5,1.5 -0.2,2 0,2.1 0.5)))`,
		},
		{
			// a triangle with a line intersecting the top vertex. Where the
			// line intersects the vertex, the line should be broken into two
			// pieces and triangulated properly.
			inputWKT:      `MULTILINESTRING((0 0,-0.1 0.5,0 1,0.5 1.2,1 1.3,1.5 1.2,2 1,2.1 0.5,2 0,1.5 -0.2,1 -0.3,0.5 -0.2,0 0),(-0.1 0.5,2.1 0.5))`,
			inputWKB:      `01050000000200000001020000000400000000000000000000000000000000000000000000000000f03f000000000000f03f00000000000000400000000000000000000000000000000000000000000000000102000000020000000000000000000000000000000000f03f0000000000000040000000000000f03f`,
			expectedEdges: `MULTILINESTRING ((1 1,2 1),(0 1,1 1),(0 0,0 1),(0 0,2 0),(2 0,2 1),(1 1,2 0),(0 0,1 1))`,
			expectedTris:  `MULTIPOLYGON (((0 1,0 0,1 1,0 1)),((1 1,0 0,2 0,1 1)),((1 1,2 0,2 1,1 1)))`,
		},
		{
			// a figure eight with a duplicate constrained line.
			inputWKT:      `MULTIPOLYGON (((0 0,0 1,1 1,1 0,0 0,0 -1,1 -1,1 0,0 0)))`,
			inputWKB:      `01060000000100000001030000000100000009000000000000000000000000000000000000000000000000000000000000000000f03f000000000000f03f000000000000f03f000000000000f03f0000000000000000000000000000000000000000000000000000000000000000000000000000f0bf000000000000f03f000000000000f0bf000000000000f03f000000000000000000000000000000000000000000000000`,
			expectedEdges: `MULTILINESTRING ((0 1,1 1),(0 0,0 1),(0 -1,0 0),(0 -1,1 -1),(1 -1,1 0),(1 0,1 1),(0 1,1 0),(0 0,1 0),(0 0,1 -1))`,
			expectedTris:  `MULTIPOLYGON (((0 1,0 0,1 0,0 1)),((0 1,1 0,1 1,0 1)),((0 -1,1 -1,0 0,0 -1)),((0 0,1 -1,1 0,0 0)))`,
		},
		{
			// A constraint line that overlaps with another edge
			inputWKT:      `MULTIPOLYGON (((0 0,1 1,2 1,3 0,3 1,0 1,0 0)))`,
			inputWKB:      `0106000000010000000103000000010000000700000000000000000000000000000000000000000000000000f03f000000000000f03f0000000000000040000000000000f03f000000000000084000000000000000000000000000000840000000000000f03f0000000000000000000000000000f03f00000000000000000000000000000000`,
			expectedEdges: `MULTILINESTRING ((2 1,3 1),(1 1,2 1),(0 1,1 1),(0 0,0 1),(0 0,3 0),(3 0,3 1),(2 1,3 0),(0 0,2 1),(0 0,1 1))`,
			expectedTris:  `MULTIPOLYGON (((0 1,0 0,1 1,0 1)),((1 1,0 0,2 1,1 1)),((2 1,0 0,3 0,2 1)),((2 1,3 0,3 1,2 1)))`,
		},
		{
			// bow-tie
			inputWKT:      `MULTIPOLYGON (((0 0,1 1,1 0,0 1,0 0)))`,
			inputWKB:      `0106000000010000000103000000010000000500000000000000000000000000000000000000000000000000f03f000000000000f03f000000000000f03f00000000000000000000000000000000000000000000f03f00000000000000000000000000000000`,
			expectedEdges: `MULTILINESTRING ((0 1,1 1),(0 0,0 1),(0 0,1 0),(1 0,1 1),(0.5 0.5,1 0),(0.5 0.5,1 1),(0 1,0.5 0.5),(0 0,0.5 0.5))`,
			expectedTris:  `MULTIPOLYGON (((0 1,0 0,0.5 0.5,0 1)),((0 1,0.5 0.5,1 1,0 1)),((1 1,0.5 0.5,1 0,1 1)),((0 0,1 0,0.5 0.5,0 0)))`,
		},
		{
			// Bow-tie with four sided concave polygon on left and triangle on 
			// right. Should break into two polygons
			inputWKT:    `POLYGON ((0 0, 0.2 0.3, 0 1, 2 0, 2 1, 0 0))`,
			inputWKB:    `01030000000100000006000000000000000000000000000000000000009a9999999999c93f333333333333d33f0000000000000000000000000000f03f000000000000004000000000000000000000000000000040000000000000f03f00000000000000000000000000000000`,
			expectedEdges: `MULTILINESTRING ((0 1,2 1),(0 0,0 1),(0 0,2 0),(2 0,2 1),(1 0.5,2 0),(1 0.5,2 1),(0 1,1 0.5),(0.2 0.3,1 0.5),(0 1,0.2 0.3),(0 0,0.2 0.3),(0 0,1 0.5))`,
			expectedTris:  `MULTIPOLYGON (((0 1,0 0,0.2 0.3,0 1)),((0 1,0.2 0.3,1 0.5,0 1)),((0 1,1 0.5,2 1,0 1)),((2 1,1 0.5,2 0,2 1)),((0 0,2 0,1 0.5,0 0)),((0 0,1 0.5,0.2 0.3,0 0)))`,
		},
		{
			// Complex multipolygon w/ erroneous overlap
			inputWKT:    `MULTIPOLYGON(((40 40,20 45,28 10,40 40)),((20 35,10 30,10 10,30 5,45 20,20 35)))`,
			inputWKB: `0106000000020000000103000000010000000400000000000000000044400000000000004440000000000000344000000000008046400000000000003c40000000000000244000000000000044400000000000004440010300000001000000060000000000000000003440000000000080414000000000000024400000000000003e40000000000000244000000000000024400000000000003e4000000000000014400000000000804640000000000000344000000000000034400000000000804140`,
			expectedEdges: `MULTILINESTRING ((20 45,40 40),(10 30,20 45),(10 10,10 30),(10 10,30 5),(30 5,45 20),(40 40,45 20),(34.516129032258064 26.29032258064516,45 20),(34.516129032258064 26.29032258064516,40 40),(22.649006622516556 33.41059602649007,34.516129032258064 26.29032258064516),(22.649006622516556 33.41059602649007,40 40),(20 45,22.649006622516556 33.41059602649007),(20 35,22.649006622516556 33.41059602649007),(20 35,20 45),(10 30,20 35),(20 35,28 10),(10 30,28 10),(10 10,28 10),(28 10,30 5),(28 10,45 20),(28 10,34.516129032258064 26.29032258064516),(22.649006622516556 33.41059602649007,28 10))`,
			expectedTris:  `MULTIPOLYGON (((10 30,10 10,28 10,10 30)),((10 30,28 10,20 35,10 30)),((10 30,20 35,20 45,10 30)),((20 45,20 35,22.649006622516556 33.41059602649007,20 45)),((20 45,22.649006622516556 33.41059602649007,40 40,20 45)),((40 40,22.649006622516556 33.41059602649007,34.516129032258064 26.29032258064516,40 40)),((40 40,34.516129032258064 26.29032258064516,45 20,40 40)),((30 5,45 20,28 10,30 5)),((30 5,28 10,10 10,30 5)),((28 10,45 20,34.516129032258064 26.29032258064516,28 10)),((28 10,34.516129032258064 26.29032258064516,22.649006622516556 33.41059602649007,28 10)),((28 10,22.649006622516556 33.41059602649007,20 35,28 10)))`,
		},
		// this test is giving correct results, but due to map ordering it is giving different results sometimes. We'll need to modify the routine to use something other than map ordering so the results are consistent.
		/*{
			// Overlapping multipolygon w/ hole. Should produce a single 
			// polygon with a smaller hole
			inputWKT:    `MULTIPOLYGON(((40 40,20 45,28 10,40 40)),((15 35,10 30,10 10,30 5,45 20,15 35),(30 20,20 15,20 25,30 20)))`,
			inputWKB:    `0106000000020000000103000000010000000400000000000000000044400000000000004440000000000000344000000000008046400000000000003c40000000000000244000000000000044400000000000004440010300000002000000060000000000000000002e40000000000080414000000000000024400000000000003e40000000000000244000000000000024400000000000003e400000000000001440000000000080464000000000000034400000000000002e400000000000804140040000000000000000003e40000000000000344000000000000034400000000000002e40000000000000344000000000000039400000000000003e400000000000003440`,
			expectedEdges: `MULTILINESTRING ((20 45,40 40),(10 30,20 45),(10 10,10 30),(10 10,30 5),(30 5,45 20),(40 40,45 20),(34.166666666666664 25.416666666666668,45 20),(34.166666666666664 25.416666666666668,40 40),(23.225806451612904 30.887096774193548,34.166666666666664 25.416666666666668),(23.225806451612904 30.887096774193548,40 40),(20 45,23.225806451612904 30.887096774193548),(15 35,23.225806451612904 30.887096774193548),(15 35,20 45),(10 30,15 35),(15 35,20 25),(10 30,20 25),(20 15,20 25),(10 30,20 15),(10 10,20 15),(20 15,28 10),(10 10,28 10),(28 10,30 5),(28 10,45 20),(28 10,34.166666666666664 25.416666666666668),(28 10,30 20),(30 20,34.166666666666664 25.416666666666668),(25.161290322580644 22.41935483870968,30 20),(25.161290322580644 22.41935483870968,34.166666666666664 25.416666666666668),(23.225806451612904 30.887096774193548,25.161290322580644 22.41935483870968),(20 25,25.161290322580644 22.41935483870968),(20 25,23.225806451612904 30.887096774193548),(20 15,25.161290322580644 22.41935483870968),(25.161290322580644 22.41935483870968,26.153846153846153 18.076923076923077),(20 15,26.153846153846153 18.076923076923077),(26.153846153846153 18.076923076923077,28 10),(26.153846153846153 18.076923076923077,30 20))`,
			expectedTris: `MULTIPOLYGON (((10 30,10 10,20 15,10 30)),((10 30,20 15,20 25,10 30)),((10 30,20 25,15 35,10 30)),((10 30,15 35,20 45,10 30)),((20 45,15 35,23.225806451612904 30.887096774193548,20 45)),((20 45,23.225806451612904 30.887096774193548,40 40,20 45)),((40 40,23.225806451612904 30.887096774193548,34.166666666666664 25.416666666666668,40 40)),((40 40,34.166666666666664 25.416666666666668,45 20,40 40)),((30 5,45 20,28 10,30 5)),((30 5,28 10,10 10,30 5)),((10 10,28 10,20 15,10 10)),((20 15,28 10,26.153846153846153 18.076923076923077,20 15)),((20 15,26.153846153846153 18.076923076923077,25.161290322580644 22.41935483870968,20 15)),((20 15,25.161290322580644 22.41935483870968,20 25,20 15)),((20 25,25.161290322580644 22.41935483870968,23.225806451612904 30.887096774193548,20 25)),((20 25,23.225806451612904 30.887096774193548,15 35,20 25)),((23.225806451612904 30.887096774193548,25.161290322580644 22.41935483870968,34.166666666666664 25.416666666666668,23.225806451612904 30.887096774193548)),((34.166666666666664 25.416666666666668,25.161290322580644 22.41935483870968,30 20,34.166666666666664 25.416666666666668)),((34.166666666666664 25.416666666666668,30 20,28 10,34.166666666666664 25.416666666666668)),((34.166666666666664 25.416666666666668,28 10,45 20,34.166666666666664 25.416666666666668)),((28 10,30 20,26.153846153846153 18.076923076923077,28 10)),((26.153846153846153 18.076923076923077,30 20,25.161290322580644 22.41935483870968,26.153846153846153 18.076923076923077)))`,
		},*/
	}

	for i, tc := range testcases {
		tc := tc
		t.Run(strconv.FormatInt(int64(i), 10), func(t *testing.T) { fn(t, tc) })
	}
}
