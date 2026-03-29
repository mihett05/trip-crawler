import './App.css';
import TripPlanner from './components/TripPlanner';
import LanguageSwitcher from './components/LanguageSwitcher';
import { AppBar, Toolbar, Typography } from '@mui/material';

function App() {
  return (
    <>
      <AppBar
        position="static"
        sx={{
          backgroundColor: '#1976d2',
          boxShadow: '0 2px 4px rgba(0,0,0,0.1)',
          marginBottom: 4,
        }}
      >
        <Toolbar>
          <Typography
            variant="h6"
            component="div"
            sx={{
              flexGrow: 1,
              fontWeight: 600,
            }}
          >
            Trip Crawler
          </Typography>
          <LanguageSwitcher />
        </Toolbar>
      </AppBar>
      <div className="App">
        <TripPlanner />
      </div>
    </>
  );
}

export default App;
