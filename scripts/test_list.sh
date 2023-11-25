go run cmd/skyman.go service list           || exit 1
go run cmd/skyman.go endpoint list          || exit 1

go run cmd/skyman.go image list             || exit 1

go run cmd/skyman.go volume list            || exit 1
go run cmd/skyman.go volume type list       || exit 1

go run cmd/skyman.go server list            || exit 1
go run cmd/skyman.go flavor list            || exit 1
go run cmd/skyman.go compute service list   || exit 1
go run cmd/skyman.go hypervisor list        || exit 1
go run cmd/skyman.go aggregate list         || exit 1
go run cmd/skyman.go az list                || exit 1
go run cmd/skyman.go az list --tree         || exit 1

go run cmd/skyman.go router list            || exit 1
go run cmd/skyman.go network list           || exit 1
go run cmd/skyman.go port list              || exit 1

