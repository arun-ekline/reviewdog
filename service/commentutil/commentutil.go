package commentutil

import (
	"encoding/base64"
	"fmt"
	"log"
	"strings"

	"github.com/reviewdog/reviewdog"
	"github.com/reviewdog/reviewdog/proto/metacomment"
	"github.com/reviewdog/reviewdog/proto/rdf"
	"google.golang.org/protobuf/proto"
)

// `path` to `position`(Lnum for new file) to comment `body` or `fingerprint`
type PostedComments map[string]map[int][]string

// IsPosted returns true if a given comment has been posted in code review service already,
// otherwise returns false. It sees comments with same path, same position,
// and same body as same comments.
func (p PostedComments) IsPosted(c *reviewdog.Comment, lineNum int, bodyOrFingerprint string) bool {
	path := c.Result.Diagnostic.GetLocation().GetPath()
	if _, ok := p[path]; !ok {
		return false
	}
	bodies, ok := p[path][lineNum]
	if !ok {
		return false
	}
	for _, b := range bodies {
		if b == bodyOrFingerprint {
			return true
		}
	}
	return false
}

// AddPostedComment adds a posted comment.
func (p PostedComments) AddPostedComment(path string, lineNum int, bodyOrFingerprint string) {
	if _, ok := p[path]; !ok {
		p[path] = make(map[int][]string)
	}
	if _, ok := p[path][lineNum]; !ok {
		p[path][lineNum] = make([]string, 0)
	}
	p[path][lineNum] = append(p[path][lineNum], bodyOrFingerprint)
}

// DebugLog outputs posted comments as log for debugging.
func (p PostedComments) DebugLog() {
	for filename, f := range p {
		for line := range f {
			log.Printf("[debug] posted: %s:%d", filename, line)
		}
	}
}

// BodyPrefix is prefix text of comment body.
const BodyPrefix = `<sub>reported by [reviewdog](https://github.com/reviewdog/reviewdog) :dog:</sub><br>`

// MarkdownComment creates comment body markdown.
func MarkdownComment(c *reviewdog.Comment) string {
	var sb strings.Builder
	if s := severity(c); s != "" {
		sb.WriteString(s)
		sb.WriteString(" ")
	}
	if tool := toolName(c); tool != "" {
		sb.WriteString(fmt.Sprintf("**[%s]** ", tool))
	}
	if code := c.Result.Diagnostic.GetCode().GetValue(); code != "" {
		if url := c.Result.Diagnostic.GetCode().GetUrl(); url != "" {
			sb.WriteString(fmt.Sprintf("<[%s](%s)> ", code, url))
		} else {
			sb.WriteString(fmt.Sprintf("<%s> ", code))
		}
	}
	sb.WriteString(BodyPrefix)
	sb.WriteString(c.Result.Diagnostic.GetMessage())
	return sb.String()
}

func toolName(c *reviewdog.Comment) string {
	if name := c.Result.Diagnostic.GetSource().GetName(); name != "" {
		return name
	}
	return c.ToolName
}

func severity(c *reviewdog.Comment) string {
	switch c.Result.Diagnostic.GetSeverity() {
	case rdf.Severity_ERROR:
		return "🚫"
	case rdf.Severity_WARNING:
		return "⚠️"
	case rdf.Severity_INFO:
		return "📝"
	default:
		return ""
	}
}

func BuildMetaComment(fprint string, toolName string) string {
	b, _ := proto.Marshal(
		&metacomment.MetaComment{
			Fingerprint: fprint,
			SourceName:  toolName,
		},
	)
	return base64.StdEncoding.EncodeToString(b)
}

func ExtractMetaComment(body string) *metacomment.MetaComment {
	prefix := "<!-- __reviewdog__:"
	for _, line := range strings.Split(body, "\n") {
		if after, found := strings.CutPrefix(line, prefix); found {
			if metastring, foundSuffix := strings.CutSuffix(after, " -->"); foundSuffix {
				meta, err := DecodeMetaComment(metastring)
				if err != nil {
					log.Printf("failed to decode MetaComment: %v", err)
					continue
				}
				return meta
			}
		}
	}
	return nil
}

func DecodeMetaComment(metaBase64 string) (*metacomment.MetaComment, error) {
	b, err := base64.StdEncoding.DecodeString(metaBase64)
	if err != nil {
		return nil, err
	}
	meta := &metacomment.MetaComment{}
	if err := proto.Unmarshal(b, meta); err != nil {
		return nil, err
	}
	return meta, nil
}

func MetaCommentTag(fprint string, toolName string) string {
	return fmt.Sprintf("\n<!-- __reviewdog__:%s -->\n", BuildMetaComment(fprint, toolName))
}
