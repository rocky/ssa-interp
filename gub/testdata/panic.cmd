# Test that debugger handles an interpreter panic()
# Use with panic.go
set highlight off
next
next
# Should see panic icon now
bt
quit
