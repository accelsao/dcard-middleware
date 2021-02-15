# dcard-middleware
- [X] client visit times count
- [X] server timeout for each IP
- [ ] start server by golang `select` (better?)
- [X] add seqential test
- [ ] add concurrent test
- [ ] use Redis for server storing

# QA
## Why use Redis, instead of golang map?
`INCR` and `EXPIRE` are useful
