package tests

import (
	"context"
	"os"

	"github.com/cucumber/godog"
	"github.com/jmoiron/sqlx"
)

var pgDSN = os.Getenv("TESTS_PG_DSN")

func init() {
	if pgDSN == "" {
		pgDSN = "postgres://ids79:ids79@pg:5432/bruteforce"
	}
}

type notifyTest struct {
	conn *sqlx.DB
	// responseStatusCode int
	// responseBody       []byte
}

func panicOnErr(err error) {
	if err != nil {
		panic(err)
	}
}

func (test *notifyTest) startPB(ctx context.Context, sc *godog.Scenario) (context.Context, error) {
	var err error
	_ = sc
	test.conn, err = sqlx.ConnectContext(ctx, "pgx", pgDSN)
	if err != nil {
		panicOnErr(err)
	}
	test.conn.Ping()

	return ctx, nil
}

func (test *notifyTest) stopPG(ctx context.Context, sc *godog.Scenario, err error) (context.Context, error) {
	_ = sc
	_ = err
	if err = test.conn.DB.Close(); err != nil {
		panicOnErr(err)
	}
	return ctx, nil
}

/* func (test *notifyTest) iSendRequestTo(httpMethod, addr string) (err error) {
	var r *http.Response

	switch httpMethod {
	case http.MethodGet:
		r, err = http.Get(addr)
	default:
		err = fmt.Errorf("unknown method: %s", httpMethod)
	}

	if err != nil {
		return
	}
	test.responseStatusCode = r.StatusCode
	test.responseBody, err = io.ReadAll(r.Body)
	return
}

func (test *notifyTest) theResponseCodeShouldBe(code int) error {
	if test.responseStatusCode != code {
		return fmt.Errorf("unexpected status code: %d != %d", test.responseStatusCode, code)
	}
	return nil
}

func (test *notifyTest) theResponseShouldMatchText(text string) error {
	if string(test.responseBody) != text {
		return fmt.Errorf("unexpected text: %s != %s", test.responseBody, text)
	}
	return nil
}

func (test *notifyTest) iSendRequestToWithData(httpMethod, addr, contentType string,
	                                           data *messages.PickleDocString) (err error) {
	var r *http.Response

	switch httpMethod {
	case http.MethodPost:
		replacer := strings.NewReplacer("\n", "", "\t", "")
		cleanJson := replacer.Replace(data.Content)
		r, err = http.Post(addr, contentType, bytes.NewReader([]byte(cleanJson)))
	default:
		err = fmt.Errorf("unknown method: %s", httpMethod)
	}

	if err != nil {
		return
	}
	test.responseStatusCode = r.StatusCode
	test.responseBody, err = io.ReadAll(r.Body)
	return
}

func (test *notifyTest) iReceiveEventWithText(text string) error {
	return nil
}*/

func InitializeScenario(s *godog.ScenarioContext) {
	test := new(notifyTest)

	s.Before(test.startPB)

	// s.Step(`^I send "([^"]*)" request to "([^"]*)"$`, test.iSendRequestTo)
	// s.Step(`^The response code should be (\d+)$`, test.theResponseCodeShouldBe)
	// s.Step(`^The response should match text "([^"]*)"$`, test.theResponseShouldMatchText)

	// s.Step(`^I send "([^"]*)" request to "([^"]*)" with "([^"]*)" data:$`, test.iSendRequestToWithData)
	// s.Step(`^I receive event with text "([^"]*)"$`, test.iReceiveEventWithText)

	s.After(test.stopPG)
}
