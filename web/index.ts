import HLS from "hls.js"
import { Events } from "hls.js"

const videoElement = document.getElementById("video") as HTMLVideoElement

console.log("using hls.js version:", HLS.version)

const URL = "http://localhost:8080"
const INDEX_URL = URL + "/index.m3u8"

const create = () => {
    if (HLS.isSupported()) {
        const config = {
            maxLiveSyncPlaybackRate: 1.5,
        }

        const hls = new HLS(config)

        hls.on(Events.ERROR, (_event, data) => {
            console.log("have error:", data)
            if (data.fatal) {
                switch (data.type) {
                    case HLS.ErrorTypes.MEDIA_ERROR:
                        console.log("media error")
                        hls.recoverMediaError()
                        break
                    case HLS.ErrorTypes.NETWORK_ERROR:
                        console.log("network error")
                        break

                    default:
                        hls.destroy()
                        setTimeout(create, 2000)
                        break
                }
            }
        })
        hls.on(Events.MEDIA_ATTACHED, () => { console.log("hls media attached") })
        hls.on(Events.MANIFEST_PARSED, (_event, data) => {
            console.log("manifest loaded", data.audioTracks)
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
