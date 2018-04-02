tomlapse
========

tomlapse polls USC's [Tommy Cam](https://web-app.usc.edu/tommycam/) webcam image
and assembles the result into a time-lapse video.

When run, tomlapse will regularly grab the latest image and save it in the
working directory as e.g. 20060102T150405Z0700.jpg. When a new image is ready,
the last 24 hours of footage is encoded as an H.264 MP4 video at 30 frames per
second called tomlapse.mp4. tomlapse requires the [FFmpeg](https://ffmpeg.org/)
command-line tool in order to create the video.

index.html plays tomlapse.mp4 in a loop, occasionally reloading the video.
