# Test of eval
# Use with expr.go
set highlight off
next
next
next
# Should be able to see expr
eval expr + " foo "
# -2
eval -2
# 5 == 6
eval 5 == 6
# 5 < 6
eval 5 < 6
# 1 << n
eval 1 << n
## FIXME: reinstate
## 1 << 8
## eval 1 << 8
# y(
eval y(
# exprs
exprs
# eval exprs[0]
eval exprs[0]
# eval exprs[100]
eval exprs[100]
# eval exprs[-9]
eval exprs[-9]
# eval os.O_RDWR | 6
eval os.O_RDWR | 6
## FIXME REINSTATE
### eval "we have: " + exprs[5] + "."
## eval "we have: " + exprs[5] + "."
quit
