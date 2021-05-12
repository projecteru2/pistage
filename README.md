# φstage

```
name: build zero redis
jobs:
  job1:
    image: harbor.shopeemobile.com/tonic/alpine:test
    timeout: 120
    steps:
      - name: login
        run:
          - echo {{dockerusername}}
          - echo {{dockerpassword}}
        with:
          dockerusername: tonic
          dockerpassword: tonicspassword

      - name: create file
        run:
          - echo test-job1 >> job1file
    files:
      - job1file

  job2:
    image: harbor.shopeemobile.com/tonic/alpine:test
    timeout: 120
    depends_on:
      - job1
    steps:
      - name: unittests
        run:
          - echo test1 >> file1
          - echo test2 >> file2

      - name: build binary
        run:
          - env
          - ls
          - echo {{ $env.GOOS }}
        env:
          GOOS: linux
          GOARCH: amd64
        with:
          version: v1

      - name: cat file
        run:
          - cat file1
          - echo "Hello $FIRST_NAME $middle_name $Last_Name, today is Monday!"
        env:
          FIRST_NAME: Mona
          middle_name: The
          Last_Name: Octocat
    files:
      - file1
      - file2

  job3:
    image: harbor.shopeemobile.com/tonic/alpine:test
    timeout: 120
    depends_on:
      - job1
      - job2
    steps:
      - name: create cluster
        run:
          - cat file1
          - cat file2
          - cat file3
          - cat job1file

      - name: store data
        run:
          - date
```