# file: features/test.feature

# http://host.docker.internal:8081/


Feature: Email notification sending
	As API client of registration service
	In order to understand that the user was informed about registration
	I want to receive event from notifications queue

	Scenario: Add to whitelist
		When I send "GET" request to "http://host.docker.internal:8081/add-white-list/?ip=10.10.10.100&mask=25"
		Then The response code should be 200
		And The response should match text "IP was added in the white list"
		Then IP "10.10.10.100" and mask "25" exist in "whitelist"
		When I send "GET" request to "http://host.docker.internal:8081/auth/?ip=10.10.10.100&login=ids&pass=123"
		Then The response code should be 200
		And The response should match 1
		When I send "GET" request to "http://host.docker.internal:8081/auth/?ip=10.10.10.100&login=ids&pass=123"
		When I send "GET" request to "http://host.docker.internal:8081/auth/?ip=10.10.10.100&login=ids&pass=123"
		When I send "GET" request to "http://host.docker.internal:8081/auth/?ip=10.10.10.100&login=ids&pass=123"
		When I send "GET" request to "http://host.docker.internal:8081/auth/?ip=10.10.10.100&login=ids&pass=123"
		When I send "GET" request to "http://host.docker.internal:8081/auth/?ip=10.10.10.100&login=ids&pass=123"
		When I send "GET" request to "http://host.docker.internal:8081/auth/?ip=10.10.10.100&login=ids&pass=123"
		When I send "GET" request to "http://host.docker.internal:8081/auth/?ip=10.10.10.100&login=ids&pass=123"
		When I send "GET" request to "http://host.docker.internal:8081/auth/?ip=10.10.10.100&login=ids&pass=123"
		When I send "GET" request to "http://host.docker.internal:8081/auth/?ip=10.10.10.100&login=ids&pass=123"
		Then The response code should be 200
		And The response should match 1		

	Scenario: Add to blacklist
		When I send "GET" request to "http://host.docker.internal:8081/add-black-list/?ip=10.10.20.100&mask=25"
		Then The response code should be 200
		And The response should match text "IP was added in the black list"
		Then IP "10.10.20.100" and mask "25" exist in "blacklist"
		When I send "GET" request to "http://host.docker.internal:8081/auth/?ip=10.10.20.50&login=ids&pass=123"
		Then The response code should be 200
		And The response should match 0

	Scenario: Auth and reset backet
		When I send "GET" request to "http://host.docker.internal:8081/reset-bucket/?ip=10.10.30.100&login=ids&pass=123"
		Then The response code should be 200
		When I send "GET" request to "http://host.docker.internal:8081/auth/?ip=10.10.30.100&login=ids&pass=123"
		Then The response code should be 200
		And The response should match 1
		When I send "GET" request to "http://host.docker.internal:8081/auth/?ip=10.10.30.100&login=ids&pass=123"
		When I send "GET" request to "http://host.docker.internal:8081/auth/?ip=10.10.30.100&login=ids&pass=123"
		When I send "GET" request to "http://host.docker.internal:8081/auth/?ip=10.10.30.100&login=ids&pass=123"
		When I send "GET" request to "http://host.docker.internal:8081/auth/?ip=10.10.30.100&login=ids&pass=123"
		When I send "GET" request to "http://host.docker.internal:8081/auth/?ip=10.10.30.100&login=ids&pass=123"
		When I send "GET" request to "http://host.docker.internal:8081/auth/?ip=10.10.30.100&login=ids&pass=123"
		When I send "GET" request to "http://host.docker.internal:8081/auth/?ip=10.10.30.100&login=ids&pass=123"
		When I send "GET" request to "http://host.docker.internal:8081/auth/?ip=10.10.30.100&login=ids&pass=123"
		When I send "GET" request to "http://host.docker.internal:8081/auth/?ip=10.10.30.100&login=ids&pass=123"
		Then The response code should be 200
		And The response should match 0
		When I send "GET" request to "http://host.docker.internal:8081/reset-bucket/?ip=10.10.30.100&login=ids&pass=123"
		Then The response code should be 200
		When I send "GET" request to "http://host.docker.internal:8081/auth/?ip=10.10.30.100&login=ids&pass=123"
		Then The response code should be 200
		And The response should match 1