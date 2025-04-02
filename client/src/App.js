import './App.css';

import {VLCPlayer, VlCPlayerView} from 'react-native-vlc-media-player';
import Orientation from 'react-native-orientation';


function App() {
    return (
        <VLCPlayer
            style={[styles.video]}
            videoAspectRatio="16:9"
            source={{uri: "https://www.radiantmediaplayer.com/media/big-buck-bunny-360p.mp4"}}
        />
        // <VlCPlayerView
        //     autoplay={false}
        //     url="https://www.radiantmediaplayer.com/media/big-buck-bunny-360p.mp4"
        //     Orientation={Orientation}
        //     ggUrl=""
        //     showGG={true}
        //     showTitle={true}
        //     title="Big Buck Bunny"
        //     showBack={true}
        //     onLeftPress={() => {
        //     }}
        // />
    );
}

export default App;
