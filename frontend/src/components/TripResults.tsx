import React from 'react';
import { Box, CardContent, Stack, Typography, Paper, Avatar } from '@mui/material';
import {
  Timeline,
  TimelineItem,
  TimelineSeparator,
  TimelineConnector,
  TimelineContent,
  TimelineDot,
} from '@mui/lab';
import { MapContainer, TileLayer, Marker, Popup } from 'react-leaflet';
import 'leaflet/dist/leaflet.css';
import { Icon } from 'leaflet';
import { useTranslation } from 'react-i18next';

// Fix for default marker icon in Leaflet with React
declare global {
  interface Window {
    __esModule?: boolean;
  }
}

delete window.__esModule;

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
  const { t } = useTranslation(); // Добавляем использование хука перевода
  // Find the center of the map based on the route
  const getMapCenter = (): [number, number] => {
    if (tripData.route.length === 0) return [51.505, -0.09]; // Default to London if no route

    const lats = tripData.route.map((segment) => segment.coordinates[0]);
    const lngs = tripData.route.map((segment) => segment.coordinates[1]);

    const avgLat = lats.reduce((sum, lat) => sum + lat, 0) / lats.length;
    const avgLng = lngs.reduce((sum, lng) => sum + lng, 0) / lngs.length;

    return [avgLat, avgLng];
  };

  return (
    <Box sx={{ mt: 4 }}>
      <Paper
        elevation={0}
        sx={{
          p: 0,
          mb: 3,
          borderRadius: 2,
          border: '1px solid',
          borderColor: 'divider',
          overflow: 'hidden',
          boxShadow: 2,
        }}
      >
        <Stack
          direction={{ xs: 'column', md: 'row' }}
          spacing={0}
          sx={{
            bgcolor: 'primary.main',
            color: 'white',
            p: 3,
          }}
        >
          <Box sx={{ flex: 1 }}>
            <Typography variant="h5" sx={{ fontWeight: 600 }}>
              {t('tripFromTo', { origin: tripData.origin, destination: tripData.destination })}
            </Typography>
            <Typography variant="body1" sx={{ mt: 0.5, opacity: 0.9 }}>
              {t('journeyInfo', {
                days: tripData.totalDays,
                distance: tripData.totalDistance.toLocaleString(),
              })}
            </Typography>
          </Box>
          <Box
            sx={{
              display: 'flex',
              alignItems: 'center',
              justifyContent: { xs: 'flex-start', md: 'flex-end' },
              mt: { xs: 2, md: 0 },
            }}
          >
            <Box sx={{ textAlign: 'right' }}>
              <Typography variant="h6" sx={{ fontWeight: 600 }}>
                {t('destinationsCount', { count: tripData.stops.length + 2 })}
              </Typography>
              <Typography variant="body2" sx={{ opacity: 0.8 }}>
                {t('includingStops')}
              </Typography>
            </Box>
          </Box>
        </Stack>

        <CardContent sx={{ p: 3, pt: 2 }}>
          <Typography
            variant="h6"
            gutterBottom
            sx={{ fontWeight: 600, mb: 2, color: 'primary.main' }}
          >
            {t('routeDetails')}
          </Typography>

          <Timeline position="left" sx={{ p: 0, mb: 3 }}>
            {tripData.route.map((segment, index) => (
              <TimelineItem key={segment.id}>
                <TimelineSeparator>
                  <TimelineDot
                    variant={
                      index === 0 || index === tripData.route.length - 1 ? 'filled' : 'outlined'
                    }
                    sx={{
                      ...(index === 0 && { bgcolor: 'success.main' }),
                      ...(index === tripData.route.length - 1 && {
                        bgcolor: 'error.main',
                      }),
                    }}
                  />
                  {index < tripData.route.length - 1 && <TimelineConnector />}
                </TimelineSeparator>
                <TimelineContent>
                  <Paper
                    elevation={0}
                    sx={{
                      p: 2,
                      border: '1px solid',
                      borderColor: 'divider',
                      borderRadius: 2,
                      bgcolor: 'background.paper',
                    }}
                  >
                    <Stack direction="row" justifyContent="space-between" alignItems="flex-start">
                      <Box>
                        <Typography variant="h6" sx={{ fontWeight: 600 }}>
                          {segment.city}
                          {index === 0 && (
                            <Typography
                              component="span"
                              sx={{
                                ml: 1,
                                px: 1,
                                py: 0.25,
                                bgcolor: 'success.light',
                                color: 'success.contrastText',
                                borderRadius: 1,
                                fontSize: '0.75rem',
                                fontWeight: 600,
                              }}
                            >
                              {t('start')}
                            </Typography>
                          )}
                          {index === tripData.route.length - 1 && (
                            <Typography
                              component="span"
                              sx={{
                                ml: 1,
                                px: 1,
                                py: 0.25,
                                bgcolor: 'error.light',
                                color: 'error.contrastText',
                                borderRadius: 1,
                                fontSize: '0.75rem',
                                fontWeight: 600,
                              }}
                            >
                              {t('end')}
                            </Typography>
                          )}
                        </Typography>

                        <Stack spacing={0.5} sx={{ mt: 1 }}>
                          <Typography variant="body2" color="text.secondary">
                            <strong>{t('arrival')}</strong>{' '}
                            {new Date(segment.arrivalDate).toLocaleDateString(undefined, {
                              weekday: 'short',
                              year: 'numeric',
                              month: 'short',
                              day: 'numeric',
                            })}
                          </Typography>
                          <Typography variant="body2" color="text.secondary">
                            <strong>{t('departure')}</strong>{' '}
                            {new Date(segment.departureDate).toLocaleDateString(undefined, {
                              weekday: 'short',
                              year: 'numeric',
                              month: 'short',
                              day: 'numeric',
                            })}
                          </Typography>
                          <Typography variant="body2" color="text.primary" sx={{ mt: 0.5 }}>
                            <strong>
                              {t('stay', {
                                count: segment.duration,
                                plural: segment.duration !== 1 ? 's' : '',
                              })}
                            </strong>
                          </Typography>
                        </Stack>
                      </Box>

                      <Avatar
                        sx={{
                          bgcolor:
                            index === 0
                              ? 'success.main'
                              : index === tripData.route.length - 1
                                ? 'error.main'
                                : 'primary.main',
                          width: 32,
                          height: 32,
                          fontSize: '0.8rem',
                        }}
                      >
                        {index + 1}
                      </Avatar>
                    </Stack>
                  </Paper>
                </TimelineContent>
              </TimelineItem>
            ))}
          </Timeline>
        </CardContent>
      </Paper>

      {/* Map Visualization */}
      <Paper
        elevation={0}
        sx={{
          borderRadius: 2,
          border: '1px solid',
          borderColor: 'divider',
          overflow: 'hidden',
          boxShadow: 2,
        }}
      >
        <Box
          sx={{
            bgcolor: 'secondary.main',
            color: 'white',
            p: 2,
            display: 'flex',
            alignItems: 'center',
          }}
        >
          <Typography variant="h6" sx={{ fontWeight: 600 }}>
            {t('tripRouteMap')}
          </Typography>
        </Box>
        <CardContent sx={{ height: 400, p: 0 }}>
          <MapContainer center={getMapCenter()} zoom={5} style={{ height: '100%', width: '100%' }}>
            <TileLayer
              attribution='&copy; <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a> contributors'
              url="https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png"
            />

            {/* Render markers for each city in the route */}
            {tripData.route.map((segment) => (
              <Marker key={segment.id} position={segment.coordinates} icon={customIcon}>
                <Popup>
                  <div>
                    <strong>{segment.city}</strong>
                    <br />
                    {t('arrive')}: {new Date(segment.arrivalDate).toLocaleDateString()}
                    <br />
                    {t('depart')}: {new Date(segment.departureDate).toLocaleDateString()}
                  </div>
                </Popup>
              </Marker>
            ))}
          </MapContainer>
        </CardContent>
      </Paper>
    </Box>
  );
};

export default TripResults;
