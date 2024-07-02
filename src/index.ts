import Hls from "hls.js"

const videoElement = document.getElementById("video") as HTMLVideoElement

// console.log(Hls.version)

const URL = "http://localhost:8080"
const INDEX_URL = URL + "/index.m3u8"

const create = () => {
    if (Hls.isSupported()) {
        const hls = new Hls({
            maxLiveSyncPlaybackRate: 1.5,
        })

        hls.on(Hls.Events.ERROR, (_event, data) => {
            console.log("have error:", data)
            if (data.fatal) {
                hls.destroy()
                setTimeout(create, 4000)
            }
        })

        // hls.loadSource("index.m3u8")
        hls.loadSource(INDEX_URL)
        hls.attachMedia(videoElement)
        videoElement.play()
    } else if (videoElement.canPlayType('application/vnd.apple.mpegurl')) {
        fetch(INDEX_URL)
            .then(() => {
                videoElement.src = INDEX_URL
                videoElement.play()
            })
    }
}

window.addEventListener("DOMContentLoaded", create)
