# Go Minimizer

Just a fun little project that transforms go code into more idiomatic go code.


### Run it


To transform the provided test file:
```
go build main.go
./main ./data/simple.go
```

To transform any other file you can now just run:

```
./main my_file.go
```

This will produce an output file named `my_file_min.go` in the same directory.


### Origins

THis project was presented at a Women Who Go Berlin event, the presentation with some info on the background and thought process can be found here:
https://docs.google.com/presentation/d/1ybERCKIICPjGMXCZDOjlJGWh8rKoKj9ZSNzBBfG7vVI/edit?usp=sharing
