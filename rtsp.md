# Run server:
#### [WORKS]
docker run --rm -it -e MTX_RTSPTRANSPORTS=tcp -e MTX_WEBRTCADDITIONALHOSTS=127.0.0.1 -p 8554:8554 -p 1935:1935 -p 8888:8888 -p 8889:8889 -p 8890:8890/udp -p 8189:8189/udp bluenviron/mediamtx 

# Stream to server:
#### original
```
ffmpeg -re -stream_loop -1 -i file.ts -c copy -f rtsp rtsp://localhost:8554/mystream
```
#### .ts
ffmpeg -re -stream_loop -1 -i video.ts -c copy -f rtsp rtsp://localhost:8554/mystream
#### [WORKS] ffmpeg .mp4 -> rtsp 
ffmpeg -re -stream_loop -1 -i video.mp4 -c copy -f rtsp rtsp://localhost:8554/mystream

ffmpeg -re -i video.mp4 -c copy -f rtsp rtsp://localhost:8554/mystream
#### .m3u8
ffmpeg -re -stream_loop -1 -i https://synema.cxllmerichie.com/proxy/0c56f96f6f49434d110f8a0a98737443:2025020601:R01lcjFaQkF1QXFCeHBCY20weGU0WVh1am5HVzVZT0swcElWN3k2M1hja2hPVURhdlFLd2xobHluODRkd2hydFJuRVplSE9qeTA1Zmx3anhPRG4rOHc9PQ==/2/4/8/6/0/4/f0yob.mp4:hls:manifest.m3u8 -c copy -f rtsp -rtsp_transport tcp rtsp://localhost:8554/mystream

# Read from server:
#### [WORKS] ffmpeg -> rtsp -> ffplay (via rtsp)
ffplay rtsp://localhost:8554/mystream
#### ffmpeg -> rtsp -> ffplay (via http)
ffplay http://localhost:8889/mystream