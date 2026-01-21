import { AppRegistry, Platform } from 'react-native';
import App from './App';

// Register without using registerRootComponent for better web compatibility
AppRegistry.registerComponent('main', () => App);

if (Platform.OS === 'web') {
    const rootTag = document.getElementById('root');
    if (rootTag) {
        AppRegistry.runApplication('main', { rootTag });
    }
}
