name: ping

uses:
  group0:
    - dep: Sleep
      with:
        duration: 8

    - run: ping -c5 8.8.4.4

  group1:
    - dep: Sleep
      with:
        duration: 8

    - run: ping -c5 8.8.8.8
