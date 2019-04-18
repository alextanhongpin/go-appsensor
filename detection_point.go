package main

type DetectionPoint string

const (
	_ DetectionPoint = ""
	// Request Exceptions.
	RE1 = "RE1"
	RE2 = "RE2"
	RE3 = "RE3"
	RE4 = "RE4"
	RE5 = "RE5"
	RE6 = "RE6"
	RE7 = "RE7"
	RE8 = "RE8"

	// Authentication Exception.
	AE1  = "AE1"
	AE2  = "AE2"
	AE3  = "AE3"
	AE4  = "AE4"
	AE5  = "AE5"
	AE6  = "AE6"
	AE7  = "AE7"
	AE8  = "AE8"
	AE9  = "AE9"
	AE10 = "AE10"
	AE11 = "AE11"
	AE12 = "AE12"
	AE13 = "AE13"

	// Session Exception.
	SE1 = "SE1"
	SE2 = "SE2"
	SE3 = "SE3"
	SE4 = "SE4"
	SE5 = "SE5"
	SE6 = "SE6"

	// Access Control Exception.
	ACE1 = "ACE1"
	ACE2 = "ACE2"
	ACE3 = "ACE3"
	ACE4 = "ACE4"

	// Input Exception.
	IE1 = "IE1"
	IE2 = "IE2"
	IE3 = "IE3"
	IE4 = "IE4"
	IE5 = "IE5"
	IE6 = "IE6"
	IE7 = "IE7"

	// Encoding Exception.
	EE1 = "EE1"
	EE2 = "EE2"

	// Command Injection Exception.
	CIE1 = "CIE1"
	CIE2 = "CIE2"
	CIE3 = "CIE3"
	CIE4 = "CIE4"

	// File IO Exception.
	FIO1 = "FIO1"
	FIO2 = "FIO2"

	// Honey Trap.
	HT1 = "HT1"
	HT2 = "HT2"
	HT3 = "HT3"

	// User Trend Exception.
	UT1 = "UT1"
	UT2 = "UT2"
	UT3 = "UT3"
	UT4 = "UT4"

	// System Trend Exception.
	STE1 = "STE1"
	STE2 = "STE2"
	STE3 = "STE3"

	// Reputation.
	RP1 = "RP1"
	RP2 = "RP2"
	RP3 = "RP3"
	RP4 = "RP4"
)

var detectionPoints = map[DetectionPoint]string{
	// Request Exception.
	RE1: "Unexpected HTTP Command",
	RE2: "Attempt to invoke unsupported HTTP Method",
	RE3: "GET When Expecting POST",
	RE4: "POST When Expecting GET",
	RE5: "Additional/Duplicated Data in Requests",
	RE6: "Data Missing from Request",
	RE7: "Unexpected Quantity of Characters in Parameter",
	RE8: "Unexpected Type of Characters in Parameter",
	// Authentication Exception.
	AE1:  "Use of Multiple Usernames",
	AE2:  "Multiple Failed Passwords",
	AE3:  "High Rate of Login Attempts",
	AE4:  "Unexpected Quantity of Characters in Username",
	AE5:  "Unexpected Quantity of Characters in Password",
	AE6:  "Unexpected Type of Characters in Username",
	AE7:  "Unexpected Type of Characters in Password",
	AE8:  "Providing Only the Username",
	AE9:  "Providing Only the Password",
	AE10: "Additional POST Variable",
	AE11: "Missing POST Variable",
	AE12: "Utilization of Common Usernames",
	AE13: "Deviation from Normal GEO Location",
	// Session Exception.
	SE1: "Modifying Existing Cookie",
	SE2: "Adding New Cookie",
	SE3: "Deleting Existing Cookie",
	SE4: "Substituting Another User's Valid Session ID or Cookie",
	SE5: "Source Location Changes During Session",
	SE6: "Change of User Agent Mid Session",
	// Access Control Exception.
	ACE1: "Modifying URL Argument Within a GET fro Direct Object Access Attempt",
	ACE2: "Modying Parameter Within a POST for Direct Object Access Attempt",
	ACE3: "Force Browsing Attempt",
	ACE4: "Evading Presentation Access Control Through Custom POST",
	// Input Exception.
	IE1: "Cross Site Scripting Attempt",
	IE2: "Violation Of Implemented White Lists",
	IE3: "Violation of Implemented Black Lists",
	IE4: "Violation of Input Data Integrity",
	IE5: "Violation of Stored Business Data Integrity",
	IE6: "Violation of Security Log Integrity",
	IE7: "Detect Abnormal Content Output Structure",

	// Encoding Exception.
	EE1: "Double Encoded Character",
	EE2: "Unexpected Encoding Used",
	// Command Injection Exception.
	CIE1: "Blacklist Inspection for Common SQL Injection Values",
	CIE2: "Detect Abnormal Quantity of Returned Records",
	CIE3: "Null Byte Character in File Request",
	CIE4: "Carriage Return or Line Feed Character in File Request",
	// File IO Exception.
	FIO1: "Detect Large Individual File",
	FIO2: "Detect Large Number of File Uploads",
	// Honey Trap.
	HT1: "Alteration to Honey Trap Date",
	HT2: "Honey Trap Resource Requested",
	HT3: "Honey Trap Data Used",
	// User Trend Exception.
	UT1: "Irregular Use of Application",
	UT2: "Speed of Application Use",
	UT3: "Frequency of Site Use",
	UT4: "Frequency of Feature Use",
	// System Trend Exception.
	STE1: "High Number of Logouts Across the Site",
	STE2: "High Number of Logins Across the Site",
	STE3: "High Number of Same Transactions Across the Site",
	// Reputation.
	RP1: "Suspicious or Disallowed User Source Location",
	RP2: "Suspicious External User Behaviour",
	RP3: "Suspicious Client-Side Behaviour",
	RP4: "Change to Environment Threat Level",
}
