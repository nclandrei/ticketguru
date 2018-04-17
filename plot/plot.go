package plot

import (
	"fmt"
	"github.com/nclandrei/ticketguru/jira"
	"github.com/wcharczuk/go-chart"
	"github.com/wcharczuk/go-chart/drawing"
	"os"
)

const (
	graphsPath = "resources/graphs"
)

// Plot defines a standard analysis plotting function.
type Plot func(...jira.Ticket) error

// Attachments draws a stacked barchart for attachments analysis.
func Attachments(tickets ...jira.Ticket) error {
	var result map[string]float64
	var withoutCount int
	var withoutTime float64
	typeCountM := make(map[jira.AttachmentType]int)
	typeTimeM := make(map[jira.AttachmentType]float64)
	for _, ticket := range tickets {
		if ticket.TimeToClose <= 0 ||
			ticket.TimeToClose > 27000 {
			continue
		}
		if len(ticket.Fields.Attachments) == 0 {
			withoutCount++
			withoutTime += ticket.TimeToClose
			continue
		}
		for _, a := range ticket.Fields.Attachments {
			typeCountM[a.Type]++
			typeTimeM[a.Type] += ticket.TimeToClose
		}
	}
	result["Without Attachments"] = withoutTime / float64(withoutCount)
	for k, v := range typeCountM {
		var score float64
		switch k {
		case jira.CodeAttachment:
			score = typeTimeM[k] / float64(v)
			result["Code"] = score
			break
		case jira.ArchiveAttachment:
			score = typeTimeM[k] / float64(v)
			result["Archive"] = score
			break
		case jira.ImageAttachment:
			score = typeTimeM[k] / float64(v)
			result["Image"] = score
			break
		case jira.ConfigAttachment:
			score = typeTimeM[k] / float64(v)
			result["Config"] = score
			break
		case jira.TextAttachment:
			score = typeTimeM[k] / float64(v)
			result["Text"] = score
			break
		case jira.SpreadsheetAttachment:
			score = typeTimeM[k] / float64(v)
			result["Spreadsheet"] = score
			break
		default:
			score = typeTimeM[k] / float64(v)
			result["Other"] = score
			break
		}
	}
	wd, err := os.Getwd()
	if err != nil {
		return err
	}
	return barchart(
		"Presence and type of attachments analysis",
		fmt.Sprintf("%s/%s/%s", wd, graphsPath, "attachments.png"),
		result,
	)
}

// StepsToReproduce produces a barchart for presence of steps to reproduce in tickets.
func StepsToReproduce(tickets ...jira.Ticket) error {
	var withCount int
	var withSum, withoutSum float64
	for _, ticket := range tickets {
		if ticket.TimeToClose <= 0 ||
			ticket.TimeToClose > 27000 {
			continue
		}
		if ticket.HasStepsToReproduce {
			withCount++
			withSum += ticket.TimeToClose
		} else {
			withoutSum += ticket.TimeToClose
		}
	}
	wd, err := os.Getwd()
	if err != nil {
		return err
	}
	return barchart(
		"Steps To Reproduce Analysis",
		fmt.Sprintf("%s/%s/%s", wd, graphsPath, "steps_to_reproduce.png"),
		map[string]float64{
			"With Steps to Reproduce":    withSum / float64(withCount),
			"Without Steps to Reproduce": withoutSum / float64(len(tickets)-withCount),
		},
	)
}

// Stacktraces produces a barchart for presence of stacktraces in tickets.
func Stacktraces(tickets ...jira.Ticket) error {
	var withCount int
	var withSum, withoutSum float64
	for _, ticket := range tickets {
		if ticket.TimeToClose <= 0 ||
			ticket.TimeToClose > 27000 {
			continue
		}
		if ticket.HasStackTrace {
			withCount++
			withSum += ticket.TimeToClose
		} else {
			withoutSum += ticket.TimeToClose
		}
	}
	wd, err := os.Getwd()
	if err != nil {
		return err
	}
	return barchart(
		"Stack Traces Analysis",
		fmt.Sprintf("%s/%s/%s", wd, graphsPath, "stack_traces.png"),
		map[string]float64{
			"With Stack Traces":    withSum / float64(withCount),
			"Without Stack Traces": withoutSum / float64(len(tickets)-withCount),
		},
	)
}

// CommentsComplexity produces a scatter plot with trendline for comments complexity analysis.
func CommentsComplexity(tickets ...jira.Ticket) error {
	var comms []float64
	var times []float64
	for _, ticket := range tickets {
		if ticket.TimeToClose > 0 &&
			ticket.TimeToClose < 27000 &&
			ticket.CommentWordsCount > 0 {
			comms = append(comms, float64(ticket.CommentWordsCount))
			times = append(times, ticket.TimeToClose)
		}
	}
	wd, err := os.Getwd()
	if err != nil {
		return err
	}
	fmt.Printf("Number of tickets with comments: %d\n", len(times))
	filePath := fmt.Sprintf("%s/%s/%s", wd, graphsPath, "comment_complexity.png")
	return scatter(
		"Comments Complexity",
		"Time-To-Close",
		"Comments Complexity Analysis",
		filePath,
		comms,
		times,
	)
}

// FieldsComplexity produces a scatter plot with trendline for fields (i.e. summary and description) complexity analysis.
func FieldsComplexity(tickets ...jira.Ticket) error {
	var fields []float64
	var times []float64
	for _, ticket := range tickets {
		if ticket.TimeToClose > 0 &&
			ticket.TimeToClose <= 27000 &&
			ticket.SummaryDescWordsCount > 0 &&
			ticket.SummaryDescWordsCount < 1000 {
			fields = append(fields, float64(ticket.SummaryDescWordsCount))
			times = append(times, ticket.TimeToClose)
		}
	}
	wd, err := os.Getwd()
	if err != nil {
		return err
	}
	filePath := fmt.Sprintf("%s/%s/%s", wd, graphsPath, "fields_complexity.png")
	return scatter(
		"Fields Complexity",
		"Time-To-Close",
		"Fields Complexity Analysis",
		filePath,
		fields,
		times,
	)
}

// GrammarCorrectness produces a scatter plot with trendline for grammar correctness scores analysis.
func GrammarCorrectness(tickets ...jira.Ticket) error {
	var scores []float64
	var times []float64
	for _, ticket := range tickets {
		if ticket.TimeToClose > 0 &&
			ticket.TimeToClose <= 27000 &&
			ticket.GrammarCorrectness.HasScore {
			scores = append(scores, float64(ticket.GrammarCorrectness.Score))
			times = append(times, ticket.TimeToClose)
		}
	}
	wd, err := os.Getwd()
	if err != nil {
		return err
	}
	fmt.Printf("Number of tickets with grammar correctness scores: %d\n", len(times))
	filePath := fmt.Sprintf("%s/%s/%s", wd, graphsPath, "grammar_correctness.png")
	return scatter(
		"Grammar Correctness Score",
		"Time-To-Close",
		"Grammar Correctness Analysis",
		filePath,
		scores,
		times,
	)
}

// SentimentAnalysis produces a scatter plot with trendline for sentiment scores analysis.
func SentimentAnalysis(tickets ...jira.Ticket) error {
	var scores []float64
	var times []float64
	for _, ticket := range tickets {
		if ticket.TimeToClose > 0 &&
			ticket.TimeToClose <= 27000 &&
			ticket.Sentiment.HasScore {
			scores = append(scores, ticket.Sentiment.Score)
			times = append(times, ticket.TimeToClose)
		}
	}
	wd, err := os.Getwd()
	if err != nil {
		return err
	}
	fmt.Printf("Number of tickets with sentiment analysis scores: %d\n", len(times))
	filePath := fmt.Sprintf("%s/%s/%s", wd, graphsPath, "sentiment_analysis.png")
	return scatter(
		"Sentiment Score",
		"Time-To-Close",
		"Sentiment Analysis",
		filePath,
		scores,
		times,
	)
}

// barchart computes and saves a barchart given a variadic number of bars.
func barchart(title, filepath string, vals map[string]float64) error {
	var bars []chart.Value
	for k, v := range vals {
		bars = append(bars, chart.Value{
			Label: k,
			Value: v,
		})
	}
	sbc := chart.BarChart{
		Title: title,
		TitleStyle: chart.Style{
			Show: true,
			Padding: chart.Box{
				Bottom: 20,
			},
		},
		Background: chart.Style{
			Padding: chart.Box{
				Top: 40,
			},
		},
		Height:   512,
		BarWidth: 60,
		XAxis: chart.Style{
			Show: true,
		},
		YAxis: chart.YAxis{
			Name: "Time-To-Close",
			NameStyle: chart.Style{
				Show: true,
			},
			Style: chart.Style{
				Show: true,
			},
		},
		Bars: bars,
	}

	file, err := os.Create(filepath)
	if err != nil {
		return err
	}

	return sbc.Render(chart.PNG, file)
}

func scatter(xAxis, yAxis, title, filepath string, xs []float64, ys []float64) error {
	viridisByY := func(xr, yr chart.Range, index int, x, y float64) drawing.Color {
		return chart.Viridis(y, yr.GetMin(), yr.GetMax())
	}

	s := chart.Chart{
		XAxis: chart.XAxis{
			Name:      xAxis,
			NameStyle: chart.Style{Show: true},
			Style:     chart.Style{Show: true},
		},
		YAxis: chart.YAxis{
			Name:      yAxis,
			NameStyle: chart.Style{Show: true},
			Style:     chart.Style{Show: true},
		},
		Title: title,
		TitleStyle: chart.Style{
			Show: true,
		},
		Series: []chart.Series{
			chart.ContinuousSeries{
				Style: chart.Style{
					Show:             true,
					StrokeWidth:      chart.Disabled,
					DotWidth:         5,
					DotColorProvider: viridisByY,
				},
				XValues: xs,
				YValues: ys,
			},
		},
	}

	file, err := os.Create(filepath)
	if err != nil {
		return err
	}

	return s.Render(chart.PNG, file)
}
