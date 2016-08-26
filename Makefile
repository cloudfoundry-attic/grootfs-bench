all:
		GOOS=linux go build .

###### Help ###################################################################

help:
		@echo '    all ................................. builds the grootfs-bench cli'
		@echo '    test ................................ runs tests locally'


###### Testing ################################################################

test:
	ginkgo -r -p -race .

