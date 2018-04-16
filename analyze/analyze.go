package analyze

import (
	"regexp"
	"strings"
	"time"

	"github.com/nclandrei/ticketguru/jira"
)

// WordinessAnalysis returns wordiness of a field (summary/comment/description) and time-to-complete (in hours).
func WordinessAnalysis(issues []jira.Issue, field string) ([]float64, []float64) {
	var wordCountSlice []float64
	var timeDiffs []float64
	for _, issue := range issues {
		timeDiff := timeToResolve(issue)
		if timeDiff > -1 && isIssueHighPriority(issue) {
			switch field {
			case "description":
				wordCountSlice = append(wordCountSlice, float64(calculateNumberOfWords(issue.Fields.Description)))
				break
			case "summary":
				wordCountSlice = append(wordCountSlice, float64(calculateNumberOfWords(issue.Fields.Summary)))
				break
			case "comment":
				wc := 0
				for _, comment := range issue.Fields.Comments.Comments {
					wc += calculateNumberOfWords(comment.Body)
				}
				wordCountSlice = append(wordCountSlice, float64(wc))
				break
			}
			timeDiffs = append(timeDiffs, timeDiff)
		}
	}
	return wordCountSlice, timeDiffs
}

// AttachmentsAnalysis returns time-to-complete (in hours) for all issues with and without attachments.
func AttachmentsAnalysis(issues []jira.Issue) ([]float64, []float64) {
	var withAttchTimeDiffs []float64
	var withoutAttchTimeDiffs []float64
	for _, issue := range issues {
		timeDiff := timeToResolve(issue)
		if timeDiff > -1 && isIssueHighPriority(issue) {
			if len(issue.Fields.Attachments) > 0 {
				withAttchTimeDiffs = append(withAttchTimeDiffs, timeDiff)
			} else {
				withoutAttchTimeDiffs = append(withoutAttchTimeDiffs, timeDiff)
			}
		}
	}
	return withAttchTimeDiffs, withoutAttchTimeDiffs
}

// SentimentScoreAnalysis returns time-to-complete and sentiment scores for input issues.
func SentimentScoreAnalysis(issues []jira.Issue) ([]float64, []float64) {
	var scores []float64
	var timeDiffs []float64
	for _, issue := range issues {
		timeDiff := timeToResolve(issue)
		if timeDiff > -1 && isIssueHighPriority(issue) && issue.Sentiment.HasScore {
			scores = append(scores, float64(issue.Sentiment.Score))
			timeDiffs = append(timeDiffs, timeDiff)
		}
	}
	return scores, timeDiffs
}

// HasStepsToReproduce returns whether an issue has steps to reproduce or not inside either
// description or any of the comments.
func HasStepsToReproduce(issue jira.Issue) (bool, error) {
	expr := `(\n(\s*)\*(.*)){2,}`
	contains, err := containsRegex(issue.Fields.Description, expr)
	if err != nil {
		return false, err
	}
	if contains {
		return true, nil
	}
	for _, comment := range issue.Fields.Comments.Comments {
		contains, err = containsRegex(comment.Body, expr)
		if err != nil {
			return false, err
		}
		if contains {
			return true, nil
		}
	}
	return false, nil
}

// HasStackTrace returns whether an issue has a stack trace attached either inside the description
// or any of the comments.
func HasStackTrace(issue jira.Issue) (bool, error) {
	expr := `^.+Exception[^\n]+\n(\s*at.+\s*\n)+`
	contains, err := containsRegex(issue.Fields.Description, expr)
	if err != nil {
		return false, err
	}
	if contains {
		return true, nil
	}
	for _, comment := range issue.Fields.Comments.Comments {
		contains, err = containsRegex(comment.Body, expr)
		if err != nil {
			return false, err
		}
		if contains {
			return true, nil
		}
	}
	return false, nil
}

func containsRegex(s, expr string) (bool, error) {
	regex, err := regexp.Compile(expr)
	if err != nil {
		return false, err
	}
	return regex.FindStringIndex(s) != nil, nil
}

// calculateNumberOfWords returns the number of words in a string.
func calculateNumberOfWords(s string) int {
	wordCount := 0
	lines := strings.Split(s, "\n")
	for _, line := range lines {
		wordCount += len(strings.Split(strings.TrimSpace(line), " "))
	}
	return wordCount
}

// GetAttachmentType returns the attachment type based on the file extension.
func getAttachmentType(filename string) jira.AttachmentType {
	extIndex := strings.LastIndex(filename, ".")
	ext := filename[(extIndex + 1):]
	switch ext {
	case "md":
		return jira.Text
	case "txt":
		return jira.Text
	case "pdf":
		return jira.Text
	case "png":
		return jira.Image
	case "jpg":
		return jira.Image
	case "jpeg":
		return jira.Image
	case "gif":
		return jira.Image
	case "bmp":
		return jira.Image
	case "mp4":
		return jira.Video
	case "avi":
		return jira.Video
	case "mkv":
		return jira.Video
	default:
		return jira.Code
	}
}

// concatAndRemoveNewLines takes a variadic number of strings and returns a concatenated form with
// all of them having newlines replaced by whitespaces.
func concatAndRemoveNewlines(strs ...string) (string, error) {
	var strBuilder strings.Builder
	for _, str := range strs {
		str = strings.Replace(str, "\n", " ", -1)
		_, err := strBuilder.WriteString(str)
		if err != nil {
			return "", err
		}
		_, err = strBuilder.WriteRune(' ')
		if err != nil {
			return "", err
		}
	}
	return strBuilder.String(), nil
}

// ConcatenateComments returns a string containing all the comment bodies concatenated.
func concatenateComments(issue jira.Issue) (string, error) {
	var builder strings.Builder
	for _, comment := range issue.Fields.Comments.Comments {
		if _, err := builder.WriteString(comment.Body); err != nil {
			return "", err
		}
	}
	return builder.String(), nil
}

// calculateTimeDifference calculates the duration in hours between 2 different timestamps.
func calculateTimeDifference(t1, t2 jira.Time) float64 {
	return time.Time(t1).Sub(time.Time(t2)).Hours()
}

// isIssueHighPriority checks whether an issue has priority ID either 1 or 2 (i.e. Critical or Major).
func isIssueHighPriority(issue jira.Issue) bool {
	return issue.Fields.Priority.ID == "1" || issue.Fields.Priority.ID == "2"
}

// timeToResolve, given an issue, returns how much time it took to close that issue.
func timeToResolve(issue jira.Issue) float64 {
	for _, history := range issue.Changelog.Histories {
		for _, item := range history.Items {
			if item.Field == "status" && item.FromString == "Open" && item.ToString == "Closed" {
				return calculateTimeDifference(history.Created, issue.Fields.Created)
			}
		}
	}
	return -1
}
