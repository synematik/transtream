ffmpeg -i "https://synema.cxllmerichie.com/proxy/f2c77277c3ae531faac9c32d2c04863d:2025020822:R01lcjFaQkF1QXFCeHBCY20weGU0WVh1am5HVzVZT0swcElWN3k2M1hja2hPVURhdlFLd2xobHluODRkd2hydGdnRjhYVGZDZmlIYUYyWjU2eVRSZ0E9PQ==/2/4/8/5/0/7/2el8n.mp4:hls:manifest.m3u8" \
-map 0:v \
-c:v libvpx-vp9 \
-s 640x360 -keyint_min 30 -g 30 \
-f webm_chunk \
-header webm_live_video_360.hdr \
-chunk_start_index 1 \
"webm_live_video_360_%d.chk" \
-map 0:a \
-c:a libvorbis \
-b:a 128k \
-f webm_chunk \
-header webm_live_audio_128.hdr \
-chunk_start_index 1 \
-audio_chunk_duration 1000 \
"webm_live_audio_128_%d.chk"