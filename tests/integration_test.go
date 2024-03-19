package tests

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/netip"
	"os"
	"time"

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
	conn               *sqlx.DB
	responseStatusCode int
	responseBody       []byte
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

func (test *notifyTest) iSendRequestTo(httpMethod, addr string) (err error) {
	if httpMethod != http.MethodGet {
		err = fmt.Errorf(" unknown method: %s", httpMethod)
		return
	}
	r, err := http.Get(addr)
	if err != nil {
		return
	}
	defer r.Body.Close()
	test.responseStatusCode = r.StatusCode
	test.responseBody, err = io.ReadAll(r.Body)
	return
}

func (test *notifyTest) theResponseCodeShouldBe(code int) error {
	if test.responseStatusCode != code {
		return fmt.Errorf(" unexpected status code: %d != %d", test.responseStatusCode, code)
	}
	return nil
}

func (test *notifyTest) theResponseShouldMatchText(text string) error {
	if string(test.responseBody) != text {
		return fmt.Errorf(" unexpected text: %s != %s", test.responseBody, text)
	}
	return nil
}

func (test *notifyTest) theResponseShouldMatch(answer int) error {
	if int(test.responseBody[0]) != answer {
		return fmt.Errorf(" unexpected answer:  %d != %d", test.responseBody[0], answer)
	}
	return nil
}

func (test *notifyTest) existInWhiteBlacklist(ip, mask, list string) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
	defer cancel()
	ipMask := ip + "/" + mask
	ipNet, err := netip.ParsePrefix(ipMask)
	if err != nil {
		return err
	}
	var query string
	if list == "whitelist" {
		query = `select * from whitelist where ip = $1`
	} else if list == "blacklist" {
		query = `select * from blacklist where ip = $1`
	}
	rows, err := test.conn.QueryContext(ctx, query, ipNet.Masked())
	if err != nil {
		return err
	}
	if rows.Next() {
		var ip string
		err := rows.Scan(&ip)
		if err != nil {
			return err
		}
		ipMask = ipNet.Masked().String()
		if ipMask != ip {
			return fmt.Errorf(" unexpected data: %v != %v", ipMask, ip)
		}
	} else {
		return fmt.Errorf(" data was not found in the %s", list)
	}
	return nil
}

func InitializeScenario(s *godog.ScenarioContext) {
	test := new(notifyTest)

	s.Before(test.startPB)

	s.Step(`^I send "([^"]*)" request to "([^"]*)"$`, test.iSendRequestTo)
	s.Step(`^The response code should be (\d+)$`, test.theResponseCodeShouldBe)
	s.Step(`^The response should match text "([^"]*)"$`, test.theResponseShouldMatchText)
	s.Step(`^IP "(\d+\.\d+\.\d+\.\d+)" and mask "(\d+)" exist in "([^"]*)"$`, test.existInWhiteBlacklist)
	s.Step(`^The response should match ([0-1])$`, test.theResponseShouldMatch)

	s.After(test.stopPG)
}
