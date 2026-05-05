package service

import "time"

const (
	// 上传状态常量
	uploadSessionStatusPending   = "pending"
	uploadSessionStatusCompleted = "completed"

	// 分块/秒传参数
	hashTypeFullSHA256              = "full_sha256_v1"
	chunkUploadThreshold            = 16 << 20
	uploadStateExpire               = 24 * time.Hour
	partUploadURLExpire             = 15 * time.Minute
	instantPrepareExpire            = 5 * time.Minute
	instantProofChallengeCount      = 3
	instantProofSegmentSize         = 4 << 10
	uploadRateLimitTTL              = 10 * time.Minute
	uploadCompleteLockTTL           = 30 * time.Second
	uploadCompleteLockMaxRefresh    = 2 * time.Minute
	maxPendingUploadSessionsPerUser = 10

	// 限流参数
	uploadInitRatePerSecond = 10.0 / 60.0
	uploadInitBurst         = 10

	uploadSignUserRatePerSecond    = 30.0
	uploadSignUserBurst            = 60
	uploadSignSessionRatePerSecond = 20.0
	uploadSignSessionBurst         = 40

	uploadReportUserRatePerSecond    = 60.0
	uploadReportUserBurst            = 120
	uploadReportSessionRatePerSecond = 40.0
	uploadReportSessionBurst         = 80

	uploadCompleteRatePerSecond = 10.0 / 60.0
	uploadCompleteBurst         = 10

	uploadQuickCheckRatePerSecond = 30.0 / 60.0
	uploadQuickCheckBurst         = 30

	uploadInstantPrepareRatePerSecond = 20.0 / 60.0
	uploadInstantPrepareBurst         = 20

	uploadInstantRatePerSecond = 20.0 / 60.0
	uploadInstantBurst         = 20
)
