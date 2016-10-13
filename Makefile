all:
		GOOS=linux go build .
		GOOS=linux go build -o grootfs-bench-reporter ./reporter

###### Help ###################################################################

help:
		@echo '    all ................................. builds the grootfs-bench cli'
		@echo '    test ................................ runs tests locally'


###### Testing ################################################################

test:
	ginkgo -r -p -race .

