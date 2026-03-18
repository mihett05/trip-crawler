import React from 'react';
import {
  Box,
  Card,
  CardContent,
  CardHeader,
  Divider,
  List,
  ListItem,
  ListItemText,
  Stack,
  Typography,
} from '@mui/material';
import { MapContainer, TileLayer, Marker, Popup } from 'react-leaflet';
import 'leaflet/dist/leaflet.css';
import { Icon } from 'leaflet';

// Fix for default marker icon in Leaflet with React
delete (window as any).__esModule;

interface TripSegment {
  id: number;
  city: string;
  coordinates: [number, number]; // [latitude, longitude]
  arrivalDate: string; // ISO date string
  departureDate: string; // ISO date string
  duration: number; // number of days
}

interface TripDetails {
  id: number;
  route: TripSegment[];
  totalDistance: number; // in kilometers
  totalDays: number;
  origin: string;
  destination: string;
  stops: string[];
}

interface TripResultsProps {
  tripData: TripDetails;
  loading?: boolean;
}

// Create a custom icon to fix the marker issue in React
const customIcon = new Icon({
  iconUrl: 'https://cdnjs.cloudflare.com/ajax/libs/leaflet/1.7.1/images/marker-icon.png',
  shadowUrl: 'https://cdnjs.cloudflare.com/ajax/libs/leaflet/1.7.1/images/marker-shadow.png',
  iconSize: [25, 41],
  iconAnchor: [12, 41],
});

const TripResults: React.FC<TripResultsProps> = ({ tripData }) => {
  // Find the center of the map based on the route
  const getMapCenter = (): [number, number] => {
    if (tripData.route.length === 0) return [51.505, -0.09]; // Default to London if no route
    
    const lats = tripData.route.map(segment => segment.coordinates[0]);
    const lngs = tripData.route.map(segment => segment.coordinates[1]);
    
    const avgLat = lats.reduce((sum, lat) => sum + lat, 0) / lats.length;
    const avgLng = lngs.reduce((sum, lng) => sum + lng, 0) / lngs.length;
    
    return [avgLat, avgLng];
  };

  return (
    <Box sx={{ mt: 4 }}>
      <Card elevation={3}>
        <CardHeader 
          title={`Trip from ${tripData.origin} to ${tripData.destination}`}
          subheader={`${tripData.totalDays} days, ${tripData.totalDistance.toLocaleString()} km`}
          sx={{ bgcolor: 'primary.main', color: 'white' }}
        />
        <CardContent>
          <Typography variant="h6" gutterBottom>
            Route Details
          </Typography>
          
          <List>
            {tripData.route.map((segment, index) => (
              <React.Fragment key={segment.id}>
                <ListItem alignItems="flex-start">
                  <ListItemText
                    primary={
                      <Typography variant="subtitle1" fontWeight="bold">
                        {segment.city} {index === 0 ? '(Origin)' : index === tripData.route.length - 1 ? '(Destination)' : `(Stop ${index})`}
                      </Typography>
                    }
                    secondary={
                      <Stack spacing={0.5}>
                        <Typography variant="body2" color="text.secondary">
                          Arrival: {new Date(segment.arrivalDate).toLocaleDateString('en-US', { weekday: 'long', year: 'numeric', month: 'long', day: 'numeric' })}
                        </Typography>
                        <Typography variant="body2" color="text.secondary">
                          Departure: {new Date(segment.departureDate).toLocaleDateString('en-US', { weekday: 'long', year: 'numeric', month: 'long', day: 'numeric' })}
                        </Typography>
                        <Typography variant="body2" color="text.primary">
                          Stay: {segment.duration} day{segment.duration !== 1 ? 's' : ''}
                        </Typography>
                      </Stack>
                    }
                  />
                </ListItem>
                {index < tripData.route.length - 1 && <Divider />}
              </React.Fragment>
            ))}
          </List>
        </CardContent>
      </Card>

      {/* Map Visualization */}
      <Card elevation={3} sx={{ mt: 3, height: 400 }}>
        <CardHeader 
          title="Trip Route Map"
          sx={{ bgcolor: 'secondary.main', color: 'white' }}
        />
        <CardContent sx={{ height: 'calc(100% - 56px)', p: 0 }}>
          <MapContainer 
            center={getMapCenter()} 
            zoom={5} 
            style={{ height: '100%', width: '100%' }}
          >
            <TileLayer
              attribution='&copy; <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a> contributors'
              url="https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png"
            />
            
            {/* Render markers for each city in the route */}
            {tripData.route.map((segment) => (
              <Marker 
                key={segment.id} 
                position={segment.coordinates} 
                icon={customIcon}
              >
                <Popup>
                  <div>
                    <strong>{segment.city}</strong><br />
                    Arrive: {new Date(segment.arrivalDate).toLocaleDateString()}<br />
                    Depart: {new Date(segment.departureDate).toLocaleDateString()}
                  </div>
                </Popup>
              </Marker>
            ))}
          </MapContainer>
        </CardContent>
      </Card>
    </Box>
  );
};

export default TripResults;