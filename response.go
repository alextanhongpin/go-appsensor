package main

type ResponseID string

const (
	// None: No Response.
	ASR_P ResponseID = "ASR-P"
	// Silent: User unaware of application's Response.
	ASR_A = "ASR_A"
	ASR_B = "ASR_B"
	ASR_C = "ASR_C"
	ASR_N = "ASR_N"
	// Passive: Changes to user experience but nothing denied.
	ASR_D = "ASR_D"
	ASR_E = "ASR_E"
	ASR_F = "ASR_F"
	// Active: Application functionality reduced for user(s).
	ASR_G = "ASR_G"
	ASR_H = "ASR_H"
	ASR_I = "ASR_I"
	ASR_J = "ASR_J"
	ASR_K = "ASR_K"
	ASR_L = "ASR_L"
	// Intrusive: User's environment altered.
	ASR_M = "ASR_M"
)

var responses = map[ResponseID]string{
	// None.
	ASR_P: "No Response",
	// Silent.
	ASR_A: "Logging Change",
	ASR_B: "Administrator Notification",
	ASR_C: "Other Notification",
	ASR_N: "Proxy",
	// Passive.
	ASR_D: "User Status Change",
	ASR_E: "User Notification",
	ASR_F: "Timing Change",
	// Active.
	ASR_G: "Process Terminated",
	ASR_H: "Function Amended",
	ASR_I: "Function Disabled",
	ASR_J: "Account Logout",
	ASR_K: "Account Lockout",
	ASR_L: "Application Disabled",
	// Intrusive.
	ASR_M: "Collect Data from User",
}
