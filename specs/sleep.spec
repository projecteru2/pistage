name: Sleep
image: alpine
async: true
run: sleep {{ .duration }}
with:
  duration: 16
