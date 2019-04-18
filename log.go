package main

type ApplicationUser struct {
	// An application-specific end-user account username, or other user
	// identity such as email address or database key, or sometimes an IP
	// address or physical device identity; never a session identifier or
	// sensitive data; possibly an "0" for unauthenticated users.
	Username string
	// User's address, e.g. IPv4 or IPv6 address.
	Source string
	// User's client software or device identification. e.g. HTTP User
	// Agent string.
	UserAgent string
	// User's client or device fingerprint e.g. SHA1 hash or certain HTTP
	// request headers.
	Fingerprint string
}

type ApplicationDetectionPoint struct {
	// The identity assigned to the activated detection point, and could
	// include further detection point details and even host, application,
	// path, code process, logic flow and instance identifiers.
	DetectionPointID string
	// The code process where the event was detected such as the module,
	// function, subroutine, component or script name (not the URL path -
	// see "entrypoint").
	DetectionPointProcess string
	// Human readable description of detection point.
	DetectionPointDescription string
	// Human readable warning message displayed to user.
	DetectionPointMessage string
}

type DetectionPointLocation struct {
	// Host system identifier e.g. host name, IP address, device identity.
	LocationHostID string

	// Application/service identifier e.g. application name abbreviation.
	LocationApplicationID string

	// Application/service release version.
	LocationApplicationVersion string

	// Network TCP or UDP port number e.g. 443.
	LocationPort int

	// Network protocol e.g. TCP, UDP.
	LocationProtocolCommunication string

	// Application protocol or physical event descriptor e.g. FTP, key,
	// HTTP screen, SIP.
	LocationProtocolApplication string

	// Application protocol method e.g. POST, depress, mouse over, touch.
	LocationMethod string

	// Data submission identifier, e.g. URL Path, button identifier, form
	// or screen name.
	LocationEntrypoint string

	// A unique identifier used to group all events associated with a
	// single user interaction e.g. when multiple detection points are
	// activated by a single user request.
	LocationInteraction string
}

type EventClassification struct {
	// Severity level from RFC5424 - The Syslog Protocol.
	Severity string

	// An integer between 0 and 100, where 100 means certain.
	Confidence int8

	// Event assignment, e.g. Operations, Compliance.
	Owner string

	// CustomName and CustomValue. Can be used for additional use but are
	// not necessarily supported by other system.
}

type EventChronology struct {
	// Timestamp from RFC3339 when the event was detected.
	EventTimestamp string

	// A Unix time (POSIX time) in the GMT time zone designated when the
	// event was logged.
	LogTimestamp int

	// Some identifier of the relevant application event log record (there
	// should be very many more application events than detection point
	// events).
	LogID string
}
