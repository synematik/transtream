ffmpeg -re -f lavfi -i testsrc=size=640x480:rate=30 -vcodec libvpx -cpu-used 5 -deadline 1 -g 10 -error-resilient 1 -auto-alt-ref 1 -f rtp 'rtp://127.0.0.1:5000?pkt_size=1200'

ffmpeg -stream_loop -1 -re -i "video.mp4" -vcodec libvpx -c:a aac -ar 44100 -b:a 128k -ac 2 -f rtp_mpegts -flush_packets 0 -fflags +genpts "rtp://127.0.0.1:5000?pkt_size=1200"

ffmpeg -stream_loop -1 -re -i "video.mp4" -c:v h264_nvenc -preset p7 -profile:v baseline -level 4.0 -pix_fmt yuv420p -forced-idr 1 -g 1 -sc_threshold 0 -c:a aac -ar 44100 -b:a 128k -ac 2 -f rtp_mpegts -flush_packets 0 -fflags +genpts "rtp://239.255.12.42:5000?localrtpport=5005&ttl=1&pkt_size=1316"