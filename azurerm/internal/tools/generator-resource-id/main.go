package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
	"time"
	"unicode"
)

func main() {
	servicePackagePath := flag.String("path", "", "The relative path to the service package")
	name := flag.String("name", "", "The name of this Resource Type")
	id := flag.String("id", "", "An example of this Resource ID")
	showHelp := flag.Bool("help", false, "Display this message")

	flag.Parse()

	if *showHelp {
		flag.Usage()
		return
	}

	if err := run(*servicePackagePath, *name, *id); err != nil {
		panic(err)
	}
}

func run(servicePackagePath, name, id string) error {
	parsersPath := fmt.Sprintf("%s/parse", servicePackagePath)
	if err := os.Mkdir(parsersPath, 0644); !os.IsExist(err) {
		return fmt.Errorf("creating parse directory at %q: %+v", parsersPath, err)
	}
	fileName := convertToSnakeCase(name)
	if strings.HasSuffix(fileName, "_test") {
		// e.g. "webtest" in applicationInsights
		fileName += "_id"
	}
	parserFilePath := fmt.Sprintf("%s/%s.go", parsersPath, fileName)
	parserTestsFilePath := fmt.Sprintf("%s/%s_test.go", parsersPath, fileName)
	resourceId, err := NewResourceID(name, id)
	if err != nil {
		return err
	}

	generator := ResourceIdGenerator{
		ResourceId: *resourceId,
	}
	if err := goFmtAndWriteToFile(parserFilePath, generator.Code()); err != nil {
		return err
	}
	if err := goFmtAndWriteToFile(parserTestsFilePath, generator.TestCode()); err != nil {
		return err
	}

	return nil
}

func convertToSnakeCase(input string) string {
	out := make([]rune, 0)
	for _, char := range input {
		if unicode.IsUpper(char) {
			out = append(out, '_')
			out = append(out, unicode.ToLower(char))
			continue
		}

		out = append(out, char)
	}
	val := string(out)
	return strings.TrimPrefix(val, "_")
}

type ResourceIdSegment struct {
	// ArgumentName is the name which should be used when this segment is used in an Argument
	ArgumentName string

	// FieldName is the name which should be used for this segment when referenced in a Field
	FieldName string

	// SegmentKey is the Segment used for this in the Resource ID e.g. `resourceGroups`
	SegmentKey string

	// SegmentValue is the value for this segment used in the Resource ID
	SegmentValue string
}

type ResourceId struct {
	TypeName string
	IDFmt    string
	IDRaw    string

	HasResourceGroup  bool
	HasSubscriptionId bool
	Segments          []ResourceIdSegment // this has to be a slice not a map since we care about the order
}

func NewResourceID(typeName, resourceId string) (*ResourceId, error) {
	// split the string, but remove the prefix of `/` since it's an empty segment
	split := strings.Split(strings.TrimPrefix(resourceId, "/"), "/")
	if len(split)%2 != 0 {
		return nil, fmt.Errorf("segments weren't divisible by 2: %q", resourceId)
	}

	segments := make([]ResourceIdSegment, 0)
	for i := 0; i < len(split); i += 2 {
		key := split[i]
		value := split[i+1]

		// the RP shouldn't be transformed
		if key == "providers" {
			continue
		}

		var segmentBuilder = func(key, value string) ResourceIdSegment {
			var toCamelCase = func(input string) string {
				// lazy but it works
				out := make([]rune, 0)
				for i, char := range strings.Title(input) {
					if i == 0 {
						out = append(out, unicode.ToLower(char))
						continue
					}

					out = append(out, char)
				}
				return string(out)
			}

			rewritten := fmt.Sprintf("%sName", key)
			segment := ResourceIdSegment{
				FieldName:    strings.Title(rewritten),
				ArgumentName: toCamelCase(rewritten),
				SegmentKey:   key,
				SegmentValue: value,
			}

			if strings.EqualFold(key, "resourceGroups") {
				segment.FieldName = "ResourceGroup"
				segment.ArgumentName = "resourceGroup"
				return segment
			}

			if key == "subscriptions" {
				segment.FieldName = "SubscriptionId"
				segment.ArgumentName = "subscriptionId"
				return segment
			}

			if strings.HasSuffix(key, "s") {
				// TODO: in time this could be worth a series of overrides

				// handles "GallerieName" and `DataFactoriesName`
				if strings.HasSuffix(key, "ies") {
					key = strings.TrimSuffix(key, "ies")
					key = fmt.Sprintf("%sy", key)
				}

				if strings.HasSuffix(key, "s") {
					key = strings.TrimSuffix(key, "s")
				}

				if strings.EqualFold(key, typeName) {
					segment.FieldName = "Name"
					segment.ArgumentName = "name"
				} else {
					// remove {Thing}s and make that {Thing}Name
					rewritten = fmt.Sprintf("%sName", key)
					segment.FieldName = strings.Title(rewritten)
					segment.ArgumentName = toCamelCase(rewritten)
				}
			}

			return segment
		}

		segments = append(segments, segmentBuilder(key, value))
	}

	// finally build up the format string based on this information
	fmtString := resourceId
	hasResourceGroup := false
	hasSubscriptionId := false
	for _, segment := range segments {
		if strings.EqualFold(segment.SegmentKey, "subscriptions") {
			hasSubscriptionId = true
		}
		if strings.EqualFold(segment.SegmentKey, "resourceGroups") {
			hasResourceGroup = true
		}

		// has to be double-escaped since this is a fmtstring
		fmtString = strings.Replace(fmtString, segment.SegmentValue, "%s", 1)
	}

	return &ResourceId{
		IDFmt:             fmtString,
		IDRaw:             resourceId,
		HasResourceGroup:  hasResourceGroup,
		HasSubscriptionId: hasSubscriptionId,
		Segments:          segments,
		TypeName:          typeName,
	}, nil
}

type ResourceIdGenerator struct {
	ResourceId
}

func (id ResourceIdGenerator) Code() string {
	return fmt.Sprintf(`
package parse

// NOTE: this file is generated via 'go:generate' - manual changes will be overwritten

import (
	"fmt"

	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/helpers/azure"
)

%s
%s
%s
%s
`, id.codeForType(), id.codeForConstructor(), id.codeForFormatter(), id.codeForParser())
}

func (id ResourceIdGenerator) codeForType() string {
	fields := make([]string, 0)
	for _, segment := range id.Segments {
		fields = append(fields, fmt.Sprintf("\t%s\tstring", segment.FieldName))
	}
	fieldStr := strings.Join(fields, "\n")
	return fmt.Sprintf(`
type %[1]sId struct {
%[2]s
}
`, id.TypeName, fieldStr)
}

func (id ResourceIdGenerator) codeForConstructor() string {
	arguments := make([]string, 0)
	assignments := make([]string, 0)

	for _, segment := range id.Segments {
		arguments = append(arguments, segment.ArgumentName)
		assignments = append(assignments, fmt.Sprintf("\t\t%s:\t%s,", segment.FieldName, segment.ArgumentName))
	}

	argumentsStr := strings.Join(arguments, ", ")
	assignmentsStr := strings.Join(assignments, "\n")
	return fmt.Sprintf(`
func New%[1]sID(%[2]s string) %[1]sId {
	return %[1]sId{
%[3]s
	}
}
`, id.TypeName, argumentsStr, assignmentsStr)
}

func (id ResourceIdGenerator) codeForFormatter() string {
	formatKeys := make([]string, 0)
	for _, segment := range id.Segments {
		formatKeys = append(formatKeys, fmt.Sprintf("id.%s", segment.FieldName))
	}
	formatKeysString := strings.Join(formatKeys, ", ")
	return fmt.Sprintf(`
func (id %[1]sId) ID(_ string) string {
	fmtString := %[2]q
	return fmt.Sprintf(fmtString, %[3]s)
}
`, id.TypeName, id.IDFmt, formatKeysString)
}

func (id ResourceIdGenerator) codeForParser() string {
	directAssignments := make([]string, 0)
	if id.HasSubscriptionId {
		directAssignments = append(directAssignments, "\t\tSubscriptionId: id.SubscriptionID,")
	}
	if id.HasResourceGroup {
		directAssignments = append(directAssignments, "\t\tResourceGroup: id.ResourceGroup,")
	}
	directAssignmentsStr := strings.Join(directAssignments, "\n")

	parserStatements := make([]string, 0)
	for _, segment := range id.Segments {
		if strings.EqualFold(segment.SegmentKey, "subscriptions") && id.HasSubscriptionId {
			// direct assigned above
			continue
		}
		if strings.EqualFold(segment.SegmentKey, "resourceGroups") && id.HasResourceGroup {
			// direct assigned above
			continue
		}

		fmtString := "\tif resourceId.%[1]s, err = id.PopSegment(\"%[2]s\"); err != nil {\n\t\treturn nil, err\n\t}"
		parserStatements = append(parserStatements, fmt.Sprintf(fmtString, segment.FieldName, segment.SegmentKey))
	}
	parserStatementsStr := strings.Join(parserStatements, "\n")

	return fmt.Sprintf(`
func %[1]sID(input string) (*%[1]sId, error) {
	id, err := azure.ParseAzureResourceID(input)
	if err != nil {
		return nil, err
	}

	resourceId := %[1]sId{
%[2]s
	}

%[3]s

	if err := id.ValidateNoEmptySegments(input); err != nil {
		return nil, err
	}

	return &resourceId, nil
}
`, id.TypeName, directAssignmentsStr, parserStatementsStr)
}

func (id ResourceIdGenerator) TestCode() string {
	return fmt.Sprintf(`
package parse

// NOTE: this file is generated via 'go:generate' - manual changes will be overwritten

import (
	"testing"

	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/resourceid"
)

%s
%s
`, id.testCodeForFormatter(), id.testCodeForParser())
}

func (id ResourceIdGenerator) testCodeForFormatter() string {
	arguments := make([]string, 0)
	for _, segment := range id.Segments {
		arguments = append(arguments, fmt.Sprintf("%q", segment.SegmentValue))
	}
	arguementsStr := strings.Join(arguments, ", ")
	return fmt.Sprintf(`
var _ resourceid.Formatter = %[1]sId{}

func Test%[1]sIDFormatter(t *testing.T) {
	actual := New%[1]sID(%[2]s).ID("")
	expected := %[3]q
	if actual != expected {
		t.Fatalf("Expected %%q but got %%q", expected, actual)
	}
}
`, id.TypeName, arguementsStr, id.IDRaw)
}

func (id ResourceIdGenerator) testCodeForParser() string {
	testCases := make([]string, 0)
	testCases = append(testCases, `
		{
			// empty
			Input: "",
			Error: true,
		},
`)
	assignmentChecks := make([]string, 0)
	for _, segment := range id.Segments {
		testCaseFmt := `
		{
			// missing %s
			Input: %q,
			Error: true,
		},`
		// missing the key
		resourceIdToThisPointIndex := strings.Index(id.IDRaw, segment.SegmentKey)
		resourceIdToThisPoint := id.IDRaw[0:resourceIdToThisPointIndex]
		testCases = append(testCases, fmt.Sprintf(testCaseFmt, segment.FieldName, resourceIdToThisPoint))

		// missing the value
		resourceIdToThisPointIndex = strings.Index(id.IDRaw, segment.SegmentValue)
		resourceIdToThisPoint = id.IDRaw[0:resourceIdToThisPointIndex]
		testCases = append(testCases, fmt.Sprintf(testCaseFmt, fmt.Sprintf("value for %s", segment.FieldName), resourceIdToThisPoint))

		assignmentsFmt := "\t\tif actual.%[1]s != v.Expected.%[1]s {\n\t\t\tt.Fatalf(\"Expected %%q but got %%q for %[1]s\", v.Expected.%[1]s, actual.%[1]s)\n\t\t}"
		assignmentChecks = append(assignmentChecks, fmt.Sprintf(assignmentsFmt, segment.FieldName))
	}

	// add a successful test case
	expectAssignments := make([]string, 0)
	for _, segment := range id.Segments {
		expectAssignments = append(expectAssignments, fmt.Sprintf("\t\t\t\t%s:\t%q,", segment.FieldName, segment.SegmentValue))
	}
	testCases = append(testCases, fmt.Sprintf(`
		{
			// valid
			Input: "%[1]s",
			Expected: &%[2]sId{
%[3]s
			},
		},
`, id.IDRaw, id.TypeName, strings.Join(expectAssignments, "\n")))

	// add an intentionally failing lower-cased test case
	testCases = append(testCases, fmt.Sprintf(`
		{
			// upper-cased
			Input: %q,
			Error: true,
		},`, strings.ToUpper(id.IDRaw)))

	testCasesStr := strings.Join(testCases, "\n")
	assignmentCheckStr := strings.Join(assignmentChecks, "\n")
	return fmt.Sprintf(`
func Test%[1]sID(t *testing.T) {
	testData := []struct {
		Input  string
		Error  bool
		Expected *%[1]sId
	}{
%[2]s
	}

	for _, v := range testData {
		t.Logf("[DEBUG] Testing %%q", v.Input)

		actual, err := %[1]sID(v.Input)
		if err != nil {
			if v.Error {
				continue
			}

			t.Fatalf("Expect a value but got an error: %%s", err)
		}
		if v.Error {
			t.Fatal("Expect an error but didn't get one")
		}

%[3]s
	}
}
`, id.TypeName, testCasesStr, assignmentCheckStr)
}

func goFmtAndWriteToFile(filePath, fileContents string) error {
	fmt, err := GolangCodeFormatter{}.Format(fileContents)
	if err != nil {
		return err
	}

	if err := ioutil.WriteFile(filePath, []byte(*fmt), 0644); err != nil {
		return err
	}

	return nil
}

type GolangCodeFormatter struct{}

func (f GolangCodeFormatter) Format(input string) (*string, error) {
	filePath := f.randomFilePath()
	if err := f.writeContentsToFile(filePath, input); err != nil {
		return nil, fmt.Errorf("writing contents to %q: %+v", filePath, err)
	}
	defer f.deleteFileContents(filePath)

	f.runGoFmt(filePath)
	f.runGoImports(filePath)

	contents, err := f.readFileContents(filePath)
	if err != nil {
		return nil, fmt.Errorf("reading contents from %q: %+v", filePath, err)
	}

	return contents, nil
}

func (f GolangCodeFormatter) randomFilePath() string {
	time := time.Now().Unix()
	return fmt.Sprintf("%stemp-%d.go", os.TempDir(), time)
}

func (f GolangCodeFormatter) runGoFmt(filePath string) {
	cmd := exec.Command("gofmt", "-w", filePath)
	// intentionally not using these errors since the exit codes are kinda uninteresting
	_ = cmd.Start()
	_ = cmd.Wait()
}

func (f GolangCodeFormatter) runGoImports(filePath string) {
	cmd := exec.Command("goimports", "-w", filePath)
	// intentionally not using these errors since the exit codes are kinda uninteresting
	_ = cmd.Start()
	_ = cmd.Wait()
}

func (f GolangCodeFormatter) deleteFileContents(filePath string) {
	_ = os.Remove(filePath)
}

func (f GolangCodeFormatter) readFileContents(filePath string) (*string, error) {
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	contents := string(data)
	return &contents, nil
}

func (GolangCodeFormatter) writeContentsToFile(filePath, contents string) error {
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	if _, err = file.WriteString(contents); err != nil {
		return err
	}

	return nil
}
