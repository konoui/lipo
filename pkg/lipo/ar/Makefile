create-archives:
		cd testdata/ &&\
        libtool -c -static arm64-func1.o -o arm64-func1.a &&\
        libtool -c -static arm64-func1.o arm64-func2.o -o arm64-func12.a &&\
        libtool -c -static arm64-func1.o arm64-func2.o arm64-func3.o -o arm64-func123.a &&\
		libtool -c -static arm64-func1.o  amd64-func1.o -o fat-arm64-amd64-func1 &&\
		ar -rc arm64-amd64-func12.a amd64-func1.o amd64-func2.o arm64-func1.o arm64-func2.o
