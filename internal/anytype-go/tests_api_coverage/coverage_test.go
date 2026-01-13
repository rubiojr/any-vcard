package tests

import (
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// ApiDefinition represents the structure of the OpenAPI definition
type ApiDefinition struct {
	Paths map[string]map[string]interface{} `json:"paths"`
	Info  map[string]interface{}            `json:"info"`
}

// EndpointInfo stores information about an API endpoint
type EndpointInfo struct {
	Path        string
	Method      string
	Tag         string
	Summary     string
	OperationID string
	Implemented bool
}

// upperFirst capitalizes just the first letter of a string
func upperFirst(s string) string {
	if s == "" {
		return ""
	}
	r := []rune(s)
	r[0] = []rune(strings.ToUpper(string(r[0])))[0]
	return string(r)
}

// convertSnakeToPascal converts snake_case to PascalCase
func convertSnakeToPascal(s string) string {
	if s == "" {
		return ""
	}

	parts := strings.Split(s, "_")
	for i, part := range parts {
		if part != "" {
			parts[i] = upperFirst(part)
		}
	}

	return strings.Join(parts, "")
}

// TestApiCoverage analyzes how much of the API definition is covered by our SDK
func TestApiCoverage(t *testing.T) {
	// Load API definition
	apiDef, err := loadApiDefinition()
	if err != nil {
		t.Fatalf("Failed to load API definition: %v", err)
	}

	// Extract all endpoints from API definition
	endpoints := extractEndpoints(apiDef)

	// Check SDK implementation for each endpoint
	checkSDKImplementation(endpoints)

	// Calculate and display coverage statistics
	displayCoverageStats(t, endpoints)
}

// loadApiDefinition reads and parses the API definition JSON file
func loadApiDefinition() (*ApiDefinition, error) {
	// Locate the API definition file
	currentDir, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("failed to get current directory: %w", err)
	}

	apiFilePath := filepath.Join(currentDir, "api_definition.json")
	data, err := os.ReadFile(apiFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read API definition file: %w", err)
	}

	// Parse JSON
	var apiDef ApiDefinition
	if err := json.Unmarshal(data, &apiDef); err != nil {
		return nil, fmt.Errorf("failed to parse API definition JSON: %w", err)
	}

	return &apiDef, nil
}

// extractEndpoints extracts all endpoints from the API definition
func extractEndpoints(apiDef *ApiDefinition) []EndpointInfo {
	var endpoints []EndpointInfo

	// Iterate through all paths and methods
	for path, pathData := range apiDef.Paths {
		for method, methodData := range pathData {
			// Skip if not an HTTP method or if methodData is not a map
			if method == "parameters" || method == "description" || method == "summary" {
				continue
			}

			methodMap, ok := methodData.(map[string]interface{})
			if !ok {
				continue
			}

			// Extract tag, summary, operationId
			var tag, summary, operationId string
			if tags, exists := methodMap["tags"]; exists {
				if tagArr, ok := tags.([]interface{}); ok && len(tagArr) > 0 {
					tag, _ = tagArr[0].(string)
				}
			}
			if sum, exists := methodMap["summary"].(string); exists {
				summary = sum
			}
			if opid, exists := methodMap["operationId"].(string); exists {
				operationId = opid
			}

			endpoints = append(endpoints, EndpointInfo{
				Path:        path,
				Method:      strings.ToUpper(method),
				Tag:         tag,
				Summary:     summary,
				OperationID: operationId,
				Implemented: false,
			})
		}
	}

	return endpoints
}

// ClientMethodInfo stores information about a method found in the SDK
type ClientMethodInfo struct {
	Name         string
	MethodType   string
	Path         string
	ParamCount   int
	ReturnCount  int
	FuncReceiver string
}

// checkSDKImplementation determines if each API endpoint is implemented in the SDK
func checkSDKImplementation(endpoints []EndpointInfo) {
	// Directory paths for different components of the SDK
	// Using absolute path to ensure we're looking in the right place
	currentDir, err := os.Getwd()
	if err != nil {
		fmt.Printf("Warning: couldn't get current directory: %v\n", err)
		return
	}

	// Go up one level to the anytype package directory, then into client
	anytypeDir := filepath.Dir(currentDir)
	clientDir := filepath.Join(anytypeDir, "client")

	fmt.Printf("Looking for client files in: %s\n", clientDir)

	// Parse all client files and extract methods
	clientMethods, err := parseClientFiles(clientDir)
	if err != nil {
		fmt.Printf("Warning: couldn't parse client files: %v\n", err)
		return
	}

	// For each endpoint, check if there's a matching client method
	for i := range endpoints {
		endpoints[i].Implemented = findMatchingMethod(&endpoints[i], clientMethods)
	}
}

// parseClientFiles parses all Go files in the client directory and extracts method information
func parseClientFiles(clientDir string) ([]ClientMethodInfo, error) {
	var methods []ClientMethodInfo

	// Get list of client files
	clientFiles, err := os.ReadDir(clientDir)
	if err != nil {
		return nil, fmt.Errorf("couldn't read client directory: %w", err)
	}

	// Create file set for parsing
	fset := token.NewFileSet()

	// Extract methods from each file
	for _, file := range clientFiles {
		if file.IsDir() || !strings.HasSuffix(file.Name(), ".go") {
			continue
		}

		// Parse the Go file
		filePath := filepath.Join(clientDir, file.Name())
		astFile, err := parser.ParseFile(fset, filePath, nil, parser.ParseComments)
		if err != nil {
			fmt.Printf("Warning: couldn't parse %s: %v\n", file.Name(), err)
			continue
		}

		// Extract methods from the file
		fileMethods := extractMethodsFromFile(astFile)
		methods = append(methods, fileMethods...)
	}

	return methods, nil
}

// extractMethodsFromFile extracts all methods and functions from a parsed Go file
func extractMethodsFromFile(file *ast.File) []ClientMethodInfo {
	var methods []ClientMethodInfo

	// Visit all declarations in the file
	for _, decl := range file.Decls {
		// Check if it's a function declaration
		funcDecl, ok := decl.(*ast.FuncDecl)
		if !ok {
			continue
		}

		// Extract function/method information
		method := ClientMethodInfo{
			Name: funcDecl.Name.Name,
		}

		// Check if it's a method (has a receiver) or a function
		if funcDecl.Recv != nil && len(funcDecl.Recv.List) > 0 {
			method.MethodType = "method"
			// Get receiver type
			if expr, ok := funcDecl.Recv.List[0].Type.(*ast.StarExpr); ok {
				if ident, ok := expr.X.(*ast.Ident); ok {
					method.FuncReceiver = ident.Name
				}
			}
		} else {
			method.MethodType = "function"
		}

		// Count parameters (excluding receiver)
		if funcDecl.Type.Params != nil {
			method.ParamCount = len(funcDecl.Type.Params.List)
		}

		// Count return values
		if funcDecl.Type.Results != nil {
			method.ReturnCount = len(funcDecl.Type.Results.List)
		}

		// Extract HTTP path and method from code and comments
		method.Path = extractEndpointFromMethod(funcDecl)

		methods = append(methods, method)
	}

	return methods
}

// extractEndpointFromMethod extracts endpoint information from a method's body and comments
func extractEndpointFromMethod(funcDecl *ast.FuncDecl) string {
	var endpoint string

	// Strategy 1: Check comments for explicit endpoint patterns
	if funcDecl.Doc != nil {
		for _, comment := range funcDecl.Doc.List {
			// Look for path pattern in comments, e.g. "POST /spaces/{space_id}"
			pathRe := regexp.MustCompile(`(GET|POST|PUT|DELETE|PATCH)\s+([/\w{}]+)`)
			if matches := pathRe.FindStringSubmatch(comment.Text); len(matches) >= 3 {
				endpoint = matches[2]
				return endpoint // Return immediately if found in comment
			}
		}
	}

	// Strategy 2: Analyze the method body for endpoint information
	if funcDecl.Body != nil {
		// Inspect the function body for endpoint assignments or calls
		ast.Inspect(funcDecl.Body, func(n ast.Node) bool {
			// Skip nil nodes
			if n == nil {
				return true
			}

			// Look for direct URL path assignments
			if assign, ok := n.(*ast.AssignStmt); ok {
				for _, rhs := range assign.Rhs {
					// For direct string assignments like urlPath := "/auth/display_code"
					if lit, ok := rhs.(*ast.BasicLit); ok && lit.Kind == token.STRING {
						pathPattern := strings.Trim(lit.Value, "\"")
						if strings.HasPrefix(pathPattern, "/") {
							// Check if it's a valid API path
							if strings.Count(pathPattern, "/") > 0 && !strings.Contains(pathPattern, "?") {
								endpoint = pathPattern
							}
						}
					}

					// Check if the right side is a call to fmt.Sprintf
					if call, ok := rhs.(*ast.CallExpr); ok {
						if sel, ok := call.Fun.(*ast.SelectorExpr); ok {
							if ident, ok := sel.X.(*ast.Ident); ok && ident.Name == "fmt" && sel.Sel.Name == "Sprintf" {
								// Extract the format string from fmt.Sprintf first argument
								if len(call.Args) > 0 {
									if lit, ok := call.Args[0].(*ast.BasicLit); ok && lit.Kind == token.STRING {
										// Extract the path pattern from the format string
										pathPattern := strings.Trim(lit.Value, "\"")

										// Extract the path part before any format specifiers or query params
										if strings.HasPrefix(pathPattern, "/") {
											basePath := extractBasePath(pathPattern)
											if basePath != "" {
												endpoint = basePath
											}
										}
									}
								}
							}
						}
					}
				}
			}

			// Look for newRequest calls with HTTP method constants or string literals
			if call, ok := n.(*ast.CallExpr); ok {
				if sel, ok := call.Fun.(*ast.SelectorExpr); ok {
					// Check if it's a newRequest call but ignore variable name assignment
					if sel.Sel.Name == "newRequest" {
						// Typically the HTTP method is the 2nd argument (after context)
						if len(call.Args) >= 3 {
							// Try to find the endpoint variable or string in the call
							endpointArg := call.Args[2]
							// If it's a direct string literal
							if lit, ok := endpointArg.(*ast.BasicLit); ok && lit.Kind == token.STRING {
								pathPattern := strings.Trim(lit.Value, "\"")
								if pathPattern != "" && strings.HasPrefix(pathPattern, "/") {
									endpoint = pathPattern
								}
							}
						}
					}
				}
			}

			return true
		})
	}

	return endpoint
}

// extractBasePath extracts the base URL path from a format string
// e.g., "/spaces/%s/objects" -> "/spaces/{}/objects"
// e.g., "/spaces/%s/objects/%s" -> "/spaces/{}/objects/{}"
func extractBasePath(formatStr string) string {
	// Replace format specifiers with placeholder
	pathRe := regexp.MustCompile(`%[sdftvq]`)
	path := pathRe.ReplaceAllString(formatStr, "{}")

	// Remove query parameters if present
	if idx := strings.Index(path, "?"); idx > -1 {
		path = path[:idx]
	}

	return path
}

// findMatchingMethod checks if there's a client method matching the endpoint
func findMatchingMethod(endpoint *EndpointInfo, methods []ClientMethodInfo) bool {
	// Strategy 1: Direct path matching
	for _, method := range methods {
		if matchPathPatterns(endpoint.Path, method.Path) {
			return true
		}
	}

	// Strategy 2: Method name inference
	possibleNames := inferMethodNames(endpoint)
	for _, possibleName := range possibleNames {
		for _, method := range methods {
			if strings.EqualFold(method.Name, possibleName) {
				return true
			}
		}
	}

	// Strategy 3: Match by tag and operation
	for _, method := range methods {
		// Check if method belongs to the expected client type based on tag
		if endpoint.Tag != "" && method.FuncReceiver != "" {
			receiverMatchesTag := strings.EqualFold(method.FuncReceiver+"Client", upperFirst(endpoint.Tag)+"Client") ||
				strings.EqualFold(method.FuncReceiver, upperFirst(endpoint.Tag)+"Client")
			if receiverMatchesTag {
				return true
			}
		}
	}

	return false
}

// matchPathPatterns checks if an API path matches a client method path
func matchPathPatterns(apiPath, clientPath string) bool {
	if apiPath == clientPath {
		return true
	}

	// If either path is empty, no match
	if apiPath == "" || clientPath == "" {
		return false
	}

	// Normalize paths by removing version prefix for comparison
	normalizeVersionPrefix := func(path string) string {
		if strings.HasPrefix(path, "/v1/") {
			return strings.TrimPrefix(path, "/v1")
		}
		return path
	}

	// Convert paths to comparable format by normalizing path parameters
	// API: /v1/spaces/{space_id} → /spaces/:param
	// Client: /spaces/{id} → /spaces/:param
	// Client: /spaces/{} → /spaces/:param (from fmt.Sprintf extraction)
	normalizePattern := func(path string) string {
		// Remove version prefix
		path = normalizeVersionPrefix(path)
		// Replace path parameters with a standard token
		re1 := regexp.MustCompile(`\{[^}]*\}`)
		re2 := regexp.MustCompile(`\{\}`)
		path = re1.ReplaceAllString(path, ":param")
		path = re2.ReplaceAllString(path, ":param")
		return path
	}

	normalizedAPI := normalizePattern(apiPath)
	normalizedClient := normalizePattern(clientPath)

	// Direct match with normalized paths
	if normalizedAPI == normalizedClient {
		return true
	}

	// Handle additional common patterns:

	// 1. Check if paths match when ignoring trailing slashes
	trimmedAPI := strings.TrimSuffix(normalizedAPI, "/")
	trimmedClient := strings.TrimSuffix(normalizedClient, "/")
	if trimmedAPI == trimmedClient {
		return true
	}

	// 2. Check if one path is a prefix of another (for nested resources)
	// Example: /spaces/:param/objects might match /spaces/:param
	// This could be refined based on actual API structure if needed
	if strings.HasPrefix(normalizedAPI, normalizedClient+"/") ||
		strings.HasPrefix(normalizedClient, normalizedAPI+"/") {
		return true
	}

	return false
}

// inferMethodNames generates possible method names based on API endpoint
func inferMethodNames(endpoint *EndpointInfo) []string {
	var names []string

	// Use operationId if present - convert snake_case to PascalCase
	if endpoint.OperationID != "" {
		// Convert snake_case operation IDs to PascalCase
		opID := convertSnakeToPascal(endpoint.OperationID)
		names = append(names, opID)

		// Also add the original in case it's already in the right format
		if opID != endpoint.OperationID {
			names = append(names, upperFirst(endpoint.OperationID))
		}
	}

	// No path, can't infer more
	if endpoint.Path == "" {
		return names
	}

	if endpoint.Path == "/v1/search" && endpoint.Method == "POST" {
		names = append(names, "Search", "SearchGlobal")
	}

	// Convert HTTP method and path to a likely function name
	// Example: GET /v1/spaces/{space_id} → GetSpace, GetSpaceByID

	// Extract path segments, removing v1 prefix
	pathForSegments := endpoint.Path
	if strings.HasPrefix(pathForSegments, "/v1/") {
		pathForSegments = strings.TrimPrefix(pathForSegments, "/v1")
	}
	segments := strings.Split(strings.Trim(pathForSegments, "/"), "/")

	// Remove path parameters
	var cleanSegments []string
	for _, segment := range segments {
		if !strings.Contains(segment, "{") {
			cleanSegments = append(cleanSegments, segment)
		}
	}

	// Build possible method names
	if len(cleanSegments) > 0 {
		// GET /spaces → GetSpaces
		// POST /spaces → CreateSpace
		// GET /spaces/{space_id} → GetSpace
		// DELETE /spaces/{space_id} → DeleteSpace

		// Determine verb based on HTTP method
		var verb string
		switch endpoint.Method {
		case "GET":
			verb = "Get"
		case "POST":
			verb = "Create"
		case "PUT":
			verb = "Update"
		case "DELETE":
			verb = "Delete"
		case "PATCH":
			verb = "Patch"
		default:
			verb = upperFirst(strings.ToLower(endpoint.Method))
		}

		// Special case for search endpoints
		if cleanSegments[len(cleanSegments)-1] == "search" {
			names = append(names, "Search")
			// continue with standard names generation as well
		}

		// Create singular form for resource name
		resourceName := cleanSegments[len(cleanSegments)-1]
		if strings.HasSuffix(resourceName, "s") {
			// For GET collections, keep plural
			if endpoint.Method == "GET" && !strings.Contains(endpoint.Path, "{") {
				names = append(names, verb+upperFirst(resourceName))
			}

			// Add singular form
			singular := strings.TrimSuffix(resourceName, "s")
			names = append(names, verb+upperFirst(singular))
		} else {
			names = append(names, verb+upperFirst(resourceName))
		}

		// Special case for nested resources
		if len(cleanSegments) > 1 {
			// GET /spaces/{space_id}/objects → GetSpaceObjects
			parentResource := cleanSegments[len(cleanSegments)-2]
			if strings.HasSuffix(parentResource, "s") {
				parentSingular := strings.TrimSuffix(parentResource, "s")
				names = append(names, verb+upperFirst(parentSingular)+upperFirst(resourceName))
			}
		}

		// Add By<Parameter> suffix if it's a parameterized endpoint
		if strings.Contains(endpoint.Path, "{") {
			// Extract parameter names
			paramRe := regexp.MustCompile(`\{([^}]+)\}`)
			matches := paramRe.FindAllStringSubmatch(endpoint.Path, -1)

			if len(matches) > 0 {
				// GET /spaces/{space_id} → GetSpaceBySpaceID
				for _, match := range matches {
					if len(match) >= 2 {
						paramName := upperFirst(strings.Replace(match[1], "_", "", -1))
						for _, baseName := range names {
							names = append(names, baseName+"By"+paramName)

							// Also add a shortened version without the redundant part
							if strings.Contains(strings.ToLower(baseName), strings.ToLower(strings.TrimSuffix(paramName, "ID"))) {
								names = append(names, baseName+"ByID")
							}
						}
					}
				}
			}
		}

		// If endpoint has a tag, add methods with tag prefix
		if endpoint.Tag != "" {
			tagName := upperFirst(endpoint.Tag)
			for _, baseName := range names {
				if !strings.HasPrefix(baseName, tagName) {
					names = append(names, tagName+baseName)
				}
			}
		}
	}

	return names
}

// displayCoverageStats calculates and displays coverage statistics
func displayCoverageStats(t *testing.T, endpoints []EndpointInfo) {
	// Calculate overall statistics
	total := len(endpoints)
	implemented := 0

	// Count by tag
	tagStats := make(map[string]struct {
		Total       int
		Implemented int
	})

	// Count by method
	methodStats := make(map[string]struct {
		Total       int
		Implemented int
	})

	// Collect statistics
	for _, e := range endpoints {
		if e.Implemented {
			implemented++
		}

		// Update tag stats
		tagStat := tagStats[e.Tag]
		tagStat.Total++
		if e.Implemented {
			tagStat.Implemented++
		}
		tagStats[e.Tag] = tagStat

		// Update method stats
		methodStat := methodStats[e.Method]
		methodStat.Total++
		if e.Implemented {
			methodStat.Implemented++
		}
		methodStats[e.Method] = methodStat
	}

	// Calculate coverage percentage
	var coverage float64
	if total > 0 {
		coverage = float64(implemented) * 100 / float64(total)
	}

	// Display overall statistics
	t.Logf("\n=== API COVERAGE REPORT ===\n")
	t.Logf("Total Endpoints: %d", total)
	t.Logf("Implemented: %d", implemented)
	t.Logf("Not Implemented: %d", total-implemented)
	t.Logf("Coverage: %.1f%%\n", coverage)

	// Display statistics by tag
	t.Logf("=== Coverage by Tag ===")
	for tag, stats := range tagStats {
		tagCoverage := float64(stats.Implemented) * 100 / float64(stats.Total)
		tagName := tag
		if tagName == "" {
			tagName = "(no tag)"
		}
		t.Logf("%s: %.1f%% (%d/%d)", tagName, tagCoverage, stats.Implemented, stats.Total)
	}
	t.Logf("")

	// Display statistics by HTTP method
	t.Logf("=== Coverage by HTTP Method ===")
	for method, stats := range methodStats {
		methodCoverage := float64(stats.Implemented) * 100 / float64(stats.Total)
		t.Logf("%s: %.1f%% (%d/%d)", method, methodCoverage, stats.Implemented, stats.Total)
	}
	t.Logf("")

	// List unimplemented endpoints
	if total-implemented > 0 {
		t.Logf("=== Unimplemented Endpoints ===")
		for _, e := range endpoints {
			if !e.Implemented {
				opInfo := ""
				if e.OperationID != "" {
					opInfo = fmt.Sprintf(" [%s]", e.OperationID)
				}
				t.Logf("%s %s%s - %s", e.Method, e.Path, opInfo, e.Summary)
			}
		}
	}
}

// AnalyzeEndpointImplementations performs a deeper analysis of endpoint implementations
// This function could be expanded to analyze method signatures, parameter types, etc.
func AnalyzeEndpointImplementations() {
	// In a more comprehensive implementation, this function would:
	// 1. Parse Go source files in the client directory
	// 2. Extract method signatures and parameters
	// 3. Compare them against expected API endpoints and their parameters
	// 4. Report any inconsistencies or missing implementations
}
