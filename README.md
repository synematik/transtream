# Server:
#### Flags:
- vcodec libx264
  - 264 seems to be fastest?
- tune zerolatency
  - idk, should optimize mb
- preset ultrafast
    - idk, should optimize mb
- filter:v fps=60
  - manual fps control
- listen 1:
  - ffmpeg fetches on client demand
- movflags frag_keyframe+empty_moov+faststart
  - frag_keyframe
    - idk
  - empty_moov
    - seems to be unseekable stream, required for mp4
  - faststart
    - literally the faster start

### [OKK] HTTP *singleclient
ffmpeg -i "https://synema.cxllmerichie.com/proxy/6e0589342c84c1e468c6442bad7cfbf4:2025020707:R01lcjFaQkF1QXFCeHBCY20weGU0WVh1am5HVzVZT0swcElWN3k2M1hja2hPVURhdlFLd2xobHluODRkd2hydFFtS2lSRGZTTC9RQVdRRjBzNzNtanc9PQ==/2/4/8/7/3/5/brh53.mp4:hls:manifest.m3u8" -listen 1 -filter:v fps=60 -f mp4 -preset ultrafast -vcodec libx264 -tune zerolatency -movflags frag_keyframe+empty_moov+faststart http://127.0.0.1:8080

### [ERR] UDP
ffmpeg -re -i https://synema.cxllmerichie.com/proxy/6e0589342c84c1e468c6442bad7cfbf4:2025020707:R01lcjFaQkF1QXFCeHBCY20weGU0WVh1am5HVzVZT0swcElWN3k2M1hja2hPVURhdlFLd2xobHluODRkd2hydFFtS2lSRGZTTC9RQVdRRjBzNzNtanc9PQ==/2/4/8/7/3/5/brh53.mp4:hls:manifest.m3u8 -vf scale=1280:720 -vcodec libx264 -pix_fmt yuv420p -tune zerolatency -preset ultrafast -f mpegts udp://236.0.0.1:2000
