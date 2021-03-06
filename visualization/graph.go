package main

import (
	"fmt"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	pusher "github.com/pusher/pusher-http-go"
	"sorting-visualization/sorting-algorithms/Randomizer"
	"strconv"
	"sync"
	"net/http"
)

// * SortIDs: * 

const heapSortID = "heap"
const bubbleSortID = "bubble"
const insertionSortID = "insertion"
const quickSortID = "quick"
const treeSortID = "tree"
const selectionSortID = "selection"


var globalArray []int

// Pusher client
var client = pusher.Client{
	AppID:   "1119581",
	Key:     "9515c01265248c7e86e8",
	Secret:  "06516341e457fe5988fd",
	Cluster: "us2",
	Secure:  true,
}

// Structs for pusher comms
type arrayElement struct {
	Value int
	Label string
}

type swapElement struct {
	Index1 int
	Index2 int
	SortID string
}

type solvedEventElement struct {
	SortID string
}

type stats struct {
	ElapsedTime float64
	Swaps       int
	Comparisons int
	Iterations int
	SortID 		string
}

func main() {
	// Echo instance
	e := echo.New()

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// Define the HTTP routes
	e.File("/", "visualization/public/index.html")
	e.Static("/style.css", "visualization/public/style.css")
	e.Static("/app.js", "visualization/public/app.js")
	e.GET("/visualize/:n", visualize)
	e.GET("/solve", solve)

	// Start server
	e.Logger.Fatal(e.Start(":11002"))
}

func visualize(c echo.Context) error {

	n, err := strconv.Atoi(c.Param("n"))

	if err !=  nil {
		fmt.Println(err)
	}

	array := Randomizer.RandomArray(n)

	globalArray = array

	fmt.Println(globalArray)

	var waitGroup sync.WaitGroup

	for _, number := range array {
		waitGroup.Add(1)
		go addNumber(number, &waitGroup)
		waitGroup.Wait()
	}

	client.Trigger("arrayVisualization", "update", 0)

	return c.String(http.StatusOK, "Visualization done")
}

func addNumber( pNumber int, pWaitGroup *sync.WaitGroup) {
	defer pWaitGroup.Done()
	arrayValue := arrayElement{
		Value: pNumber,
		Label: "",
	}
	client.Trigger("arrayVisualization", "addNumber", arrayValue)
}

func solve(c echo.Context) error {

	bubbleArray    := make([]int, len(globalArray))
	insertionArray := make([]int, len(globalArray))
	heapArray      := make([]int, len(globalArray))
	quickArray     := make([]int, len(globalArray))
	selectionArray := make([]int, len(globalArray))
	treeArray      := make([]int, len(globalArray))

	copy(heapArray,      globalArray)
	copy(quickArray,     globalArray)
	copy(bubbleArray,    globalArray)
	copy(insertionArray, globalArray)
	copy(selectionArray, globalArray)
	copy(treeArray,      globalArray)

	client.Trigger("arrayVisualization", "startSolving", 0)

	var waitGroup sync.WaitGroup
	waitGroup.Add(6)

	//Visualizer foos here
	go AlgorithmVisualizer(treeArray,      treeSortID,      &waitGroup)
	go AlgorithmVisualizer(heapArray,      heapSortID,      &waitGroup)
	go AlgorithmVisualizer(bubbleArray,    bubbleSortID,    &waitGroup)
	go AlgorithmVisualizer(insertionArray, insertionSortID, &waitGroup)
	go AlgorithmVisualizer(selectionArray, selectionSortID, &waitGroup)
	go AlgorithmVisualizer(quickArray,     quickSortID,     &waitGroup)

	waitGroup.Wait()

	fmt.Println("terminan")

	return c.String(http.StatusOK, "Simulation begun")

}

// VisualSwap swaps two bars in visualization
func VisualSwap( pIndex1 int, pIndex2 int, pSortID string ){
	swapData := swapElement{
		Index1: pIndex1,
		Index2: pIndex2,
		SortID: pSortID,
	}
	client.Trigger("arrayVisualization", "swap", swapData)
}
// ShowStats sends the number of iterations, swaps,comparisons and elapsed time to the visualization
func ShowStats(pElapsedTime float64, pIterations int, pSwaps int, pComparisons int, pSortID string){
	statData := stats{
		ElapsedTime: pElapsedTime,
		Swaps: pSwaps,
		Comparisons: pComparisons,
		Iterations: pIterations,
		SortID: pSortID,
	}
	client.Trigger("arrayVisualization", "showStats", statData)
}
// VisualDone Indicates in the visualization when an algorithm is done
func VisualDone(pSortID string){
	client.Trigger("arrayVisualization", "solved", solvedEventElement{pSortID})
}

