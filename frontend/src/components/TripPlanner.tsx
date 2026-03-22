import React, { useState } from 'react';
import {
  Box,
  Button,
  Card,
  CardContent,
  CardActions,
  Chip,
  CircularProgress,
  FormControl,
  FormHelperText,
  InputLabel,
  MenuItem,
  Select,
  Stack,
  TextField,
  Typography,
  Paper,
  Divider,
  IconButton,
} from '@mui/material';
import { LocalizationProvider } from '@mui/x-date-pickers/LocalizationProvider';
import { AdapterDateFns } from '@mui/x-date-pickers/AdapterDateFns';
import { DatePicker } from '@mui/x-date-pickers/DatePicker';
import { Autocomplete } from '@mui/material';
import { SwapHoriz as SwapIcon } from '@mui/icons-material';
import { useQuery } from '@tanstack/react-query';
import { useTranslation } from 'react-i18next';
import {
  fetchCitySuggestions,
  convertFormDataToApiRequest,
  convertApiToTripDetails,
} from '../services/api';
import TripResults from './TripResults';
import { useCreateRoute } from '../gen';

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

interface TripFormData {
  departureCity: string;
  middleCities: string[];
  destinationCity: string;
  startDate: Date | null;
  tripDuration: number;
}

const TripPlanner: React.FC = () => {
  const { t } = useTranslation(); // Добавляем использование хука перевода
  const [formData, setFormData] = useState<TripFormData>({
    departureCity: '',
    middleCities: [],
    destinationCity: '',
    startDate: null,
    tripDuration: 7,
  });

  const [currentMiddleCity, setCurrentMiddleCity] = useState('');
  const [errors, setErrors] = useState<Record<string, string>>({});
  const [showResults, setShowResults] = useState(false);
  const [tripResult, setTripResult] = useState<TripDetails | null>(null);

  // Query for autocompleting departure city
  const { data: departureSuggestions = [] } = useQuery({
    queryKey: ['cities', formData.departureCity],
    queryFn: () => fetchCitySuggestions(formData.departureCity),
    enabled: formData.departureCity.length > 1,
  });

  // Query for autocompleting destination city
  const { data: destinationSuggestions = [] } = useQuery({
    queryKey: ['cities', formData.destinationCity],
    queryFn: () => fetchCitySuggestions(formData.destinationCity),
    enabled: formData.destinationCity.length > 1,
  });

  // Query for autocompleting middle cities
  const { data: middleCitySuggestions = [] } = useQuery({
    queryKey: ['cities', currentMiddleCity],
    queryFn: () => fetchCitySuggestions(currentMiddleCity),
    enabled: currentMiddleCity.length > 1,
  });

  // Mutation for creating trip route via API
  const generateTripMutation = useCreateRoute();

  const handleInputChange = (
    field: keyof TripFormData,
    value: string | number | Date | null | string[],
  ) => {
    setFormData((prev) => ({ ...prev, [field]: value }));

    // Clear error when field is changed
    if (errors[field]) {
      setErrors((prev) => {
        const newErrors = { ...prev };
        delete newErrors[field];
        return newErrors;
      });
    }
  };

  const handleSwapCities = () => {
    const temp = formData.departureCity;
    setFormData((prev) => ({
      ...prev,
      departureCity: prev.destinationCity,
      destinationCity: temp,
    }));
  };

  const handleAddMiddleCity = () => {
    if (currentMiddleCity.trim() && formData.middleCities.length < 3) {
      setFormData((prev) => ({
        ...prev,
        middleCities: [...prev.middleCities, currentMiddleCity],
      }));
      setCurrentMiddleCity('');
    }
  };

  const handleRemoveMiddleCity = (index: number) => {
    setFormData((prev) => ({
      ...prev,
      middleCities: prev.middleCities.filter((_, i) => i !== index),
    }));
  };

  const validateForm = (): boolean => {
    const newErrors: Record<string, string> = {};

    if (!formData.departureCity.trim()) {
      newErrors.departureCity = t('startingCityRequired');
    }

    if (!formData.destinationCity.trim()) {
      newErrors.destinationCity = t('destinationCityRequired');
    }

    if (!formData.startDate) {
      newErrors.startDate = t('startDateRequired');
    }

    if (formData.tripDuration <= 0) {
      newErrors.tripDuration = t('tripDurationMin');
    }

    setErrors(newErrors);
    return Object.keys(newErrors).length === 0;
  };

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (validateForm()) {
      const apiRequest = convertFormDataToApiRequest(formData);
      generateTripMutation.mutateAsync({ data: apiRequest }).then((resp) => {
        const tripDetails = convertApiToTripDetails(resp, formData);
        setTripResult(tripDetails);
        setShowResults(true);
      });
    }
  };

  const handleNewTrip = () => {
    setShowResults(false);
    setTripResult(null);
  };

  if (showResults && tripResult) {
    return (
      <Box sx={{ maxWidth: 1000, mx: 'auto', mt: 4, px: 2 }}>
        <Paper
          elevation={0}
          sx={{
            p: 3,
            mb: 3,
            borderRadius: 2,
            bgcolor: 'background.paper',
            boxShadow: 1,
          }}
        >
          <Typography
            variant="h4"
            component="h1"
            align="center"
            gutterBottom
            sx={{ fontWeight: 600, color: 'primary.main' }}
          >
            {t('yourTripPlan')}
          </Typography>
        </Paper>

        <TripResults tripData={tripResult} />

        <Box sx={{ textAlign: 'center', mt: 3 }}>
          <Button variant="outlined" onClick={handleNewTrip} sx={{ mt: 2, px: 4, py: 1.2 }}>
            {t('planAnotherTrip')}
          </Button>
        </Box>
      </Box>
    );
  }

  return (
    <LocalizationProvider dateAdapter={AdapterDateFns}>
      <Box sx={{ maxWidth: 1000, mx: 'auto', mt: 4, px: 2 }}>
        <Paper
          elevation={0}
          sx={{
            p: 3,
            mb: 3,
            borderRadius: 2,
            bgcolor: 'background.paper',
            boxShadow: 1,
          }}
        >
          <Typography
            variant="h4"
            component="h1"
            align="center"
            gutterBottom
            sx={{ fontWeight: 600, color: 'primary.main' }}
          >
            {t('appTitle')}
          </Typography>
          <Typography variant="body1" align="center" color="textSecondary" sx={{ mb: 1 }}>
            {t('appSubtitle')}
          </Typography>
        </Paper>

        <Card
          elevation={0}
          sx={{
            borderRadius: 2,
            border: '1px solid',
            borderColor: 'divider',
            overflow: 'visible',
            boxShadow: 2,
          }}
        >
          <CardContent sx={{ pb: 2 }}>
            <form onSubmit={handleSubmit}>
              <Box
                sx={{
                  display: 'flex',
                  flexDirection: { xs: 'column', md: 'row' },
                  alignItems: 'center',
                  gap: 2,
                  width: '100%',
                }}
              >
                {/* Departure City */}
                <Box sx={{ flex: 5, width: '100%' }}>
                  <Autocomplete
                    freeSolo
                    options={departureSuggestions}
                    value={formData.departureCity}
                    onChange={(_event, newValue) =>
                      handleInputChange('departureCity', newValue || '')
                    }
                    inputValue={formData.departureCity}
                    onInputChange={(_event, newInputValue) =>
                      handleInputChange('departureCity', newInputValue)
                    }
                    renderInput={(params) => (
                      <TextField
                        {...params}
                        label={t('from')}
                        error={!!errors.departureCity}
                        helperText={errors.departureCity || ''}
                        fullWidth
                        InputProps={{
                          ...params.InputProps,
                          sx: {
                            fontSize: '1.1rem',
                            fontWeight: 500,
                            pl: 1,
                          },
                        }}
                        InputLabelProps={{
                          sx: {
                            fontSize: '0.9rem',
                            fontWeight: 600,
                          },
                        }}
                      />
                    )}
                  />
                </Box>

                {/* Swap Button */}
                <Box
                  sx={{
                    display: 'flex',
                    justifyContent: 'center',
                    alignItems: 'center',
                    flex: 2,
                  }}
                >
                  <IconButton
                    onClick={handleSwapCities}
                    sx={{
                      bgcolor: 'primary.main',
                      color: 'white',
                      width: 40,
                      height: 40,
                      '&:hover': {
                        bgcolor: 'primary.dark',
                      },
                    }}
                  >
                    <SwapIcon />
                  </IconButton>
                </Box>

                {/* Destination City */}
                <Box sx={{ flex: 5, width: '100%' }}>
                  <Autocomplete
                    freeSolo
                    options={destinationSuggestions}
                    value={formData.destinationCity}
                    onChange={(_event, newValue) =>
                      handleInputChange('destinationCity', newValue || '')
                    }
                    inputValue={formData.destinationCity}
                    onInputChange={(_event, newInputValue) =>
                      handleInputChange('destinationCity', newInputValue)
                    }
                    renderInput={(params) => (
                      <TextField
                        {...params}
                        label={t('to')}
                        error={!!errors.destinationCity}
                        helperText={errors.destinationCity || ''}
                        fullWidth
                        InputProps={{
                          ...params.InputProps,
                          sx: {
                            fontSize: '1.1rem',
                            fontWeight: 500,
                            pl: 1,
                          },
                        }}
                        InputLabelProps={{
                          sx: {
                            fontSize: '0.9rem',
                            fontWeight: 600,
                          },
                        }}
                      />
                    )}
                  />
                </Box>
              </Box>

              {/* Middle Cities */}
              <Box sx={{ mt: 3 }}>
                <Typography variant="h6" gutterBottom sx={{ fontWeight: 600, mb: 1.5 }}>
                  {t('stopoverCities')}
                </Typography>

                <Stack
                  direction={{ xs: 'column', sm: 'row' }}
                  spacing={1}
                  alignItems="flex-start"
                  mb={1}
                >
                  <Autocomplete
                    freeSolo
                    options={middleCitySuggestions.filter(
                      (city) =>
                        !formData.middleCities.includes(city) &&
                        city !== formData.departureCity &&
                        city !== formData.destinationCity,
                    )}
                    value={currentMiddleCity}
                    onChange={(_event, newValue) => setCurrentMiddleCity(newValue || '')}
                    inputValue={currentMiddleCity}
                    onInputChange={(_event, newInputValue) => setCurrentMiddleCity(newInputValue)}
                    sx={{ flex: 1 }}
                    renderInput={(params) => (
                      <TextField
                        {...params}
                        label={t('addStopoverCity')}
                        fullWidth
                        InputProps={{
                          ...params.InputProps,
                          sx: { pl: 1 },
                        }}
                      />
                    )}
                  />
                  <Button
                    variant="outlined"
                    onClick={handleAddMiddleCity}
                    disabled={currentMiddleCity.trim() === '' || formData.middleCities.length >= 3}
                    sx={{ mt: { xs: 1, sm: 2.5 }, px: 2 }}
                  >
                    {t('add')}
                  </Button>
                </Stack>

                {formData.middleCities.length > 0 && (
                  <Stack direction="row" spacing={1} flexWrap="wrap" sx={{ mt: 1 }}>
                    {formData.middleCities.map((city, index) => (
                      <Chip
                        key={index}
                        label={city}
                        onDelete={() => handleRemoveMiddleCity(index)}
                        color="primary"
                        variant="filled"
                        sx={{ borderRadius: 1 }}
                      />
                    ))}
                  </Stack>
                )}

                <FormHelperText sx={{ ml: 0, mt: 1 }}>
                  {t('citiesAdded', { count: formData.middleCities.length })}
                </FormHelperText>
              </Box>

              <Divider sx={{ my: 3 }} />

              {/* Dates and Duration */}
              <Box
                sx={{
                  display: 'grid',
                  gridTemplateColumns: { xs: '1fr', md: '1fr 1fr' },
                  gap: 2,
                  mb: 2,
                }}
              >
                <Box>
                  <Box>
                    <DatePicker
                      label={t('departureDate')}
                      value={formData.startDate}
                      onChange={(newValue) => handleInputChange('startDate', newValue)}
                      minDate={new Date()}
                      slotProps={{
                        textField: {
                          error: !!errors.startDate,
                          helperText: errors.startDate || '',
                          fullWidth: true,
                        },
                        actionBar: {
                          actions: ['today', 'accept'],
                        },
                      }}
                      sx={{
                        width: '100%',
                        '& .MuiInputBase-input': {
                          py: 1.2,
                          px: 1.5,
                        },
                      }}
                    />
                  </Box>
                </Box>

                <Box>
                  <FormControl fullWidth error={!!errors.tripDuration}>
                    <InputLabel id="duration-select-label" sx={{ fontWeight: 600 }}>
                      {t('tripLength')}
                    </InputLabel>
                    <Select
                      labelId="duration-select-label"
                      value={formData.tripDuration}
                      label={t('tripLength')}
                      onChange={(e) => handleInputChange('tripDuration', Number(e.target.value))}
                      sx={{
                        py: 1.2,
                        px: 1.5,
                        '& .MuiTypography-root': {
                          fontWeight: 500,
                        },
                      }}
                    >
                      {[...Array(30)].map((_, i) => (
                        <MenuItem key={i + 1} value={i + 1}>
                          {t('days', { count: i + 1, plural: i + 1 !== 1 ? '' : '' })}
                        </MenuItem>
                      ))}
                    </Select>
                    {errors.tripDuration && <FormHelperText>{errors.tripDuration}</FormHelperText>}
                  </FormControl>
                </Box>
              </Box>

              {/* Submit Button */}
              <CardActions sx={{ justifyContent: 'center', pt: 1, pb: 0 }}>
                <Button
                  type="submit"
                  variant="contained"
                  size="large"
                  disabled={generateTripMutation.isPending}
                  sx={{
                    px: 6,
                    py: 1.5,
                    minWidth: 220,
                    fontSize: '1.1rem',
                    fontWeight: 600,
                    boxShadow: 2,
                    '&:hover': {
                      boxShadow: 3,
                    },
                  }}
                  startIcon={generateTripMutation.isPending ? <CircularProgress size={20} /> : null}
                >
                  {generateTripMutation.isPending ? t('planning') : t('findTrips')}
                </Button>
              </CardActions>
            </form>
          </CardContent>
        </Card>
      </Box>
    </LocalizationProvider>
  );
};

export default TripPlanner;
