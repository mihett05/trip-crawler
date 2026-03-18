import React, { useState } from 'react';
import {
  Box,
  Button,
  Card,
  CardContent,
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
} from '@mui/material';
import { LocalizationProvider } from '@mui/x-date-pickers/LocalizationProvider';
import { AdapterDateFns } from '@mui/x-date-pickers/AdapterDateFns';
import { DatePicker } from '@mui/x-date-pickers/DatePicker';
import { Autocomplete } from '@mui/material';
import { useQuery, useMutation } from '@tanstack/react-query';
import { fetchCitySuggestions, generateTripRoute } from '../services/api';
import TripResults from './TripResults';

interface TripFormData {
  departureCity: string;
  middleCities: string[];
  destinationCity: string;
  startDate: Date | null;
  tripDuration: number;
}

const TripPlanner: React.FC = () => {
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
  const [tripResult, setTripResult] = useState<any>(null);

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

  // Mutation for generating trip route
  const generateTripMutation = useMutation({
    mutationFn: generateTripRoute,
    onSuccess: (data) => {
      setTripResult(data);
      setShowResults(true);
    },
  });

  const handleInputChange = (field: keyof TripFormData, value: any) => {
    setFormData(prev => ({ ...prev, [field]: value }));
    
    // Clear error when field is changed
    if (errors[field]) {
      setErrors(prev => {
        const newErrors = { ...prev };
        delete newErrors[field];
        return newErrors;
      });
    }
  };

  const handleAddMiddleCity = () => {
    if (currentMiddleCity.trim() && formData.middleCities.length < 3) {
      setFormData(prev => ({
        ...prev,
        middleCities: [...prev.middleCities, currentMiddleCity],
      }));
      setCurrentMiddleCity('');
    }
  };

  const handleRemoveMiddleCity = (index: number) => {
    setFormData(prev => ({
      ...prev,
      middleCities: prev.middleCities.filter((_, i) => i !== index),
    }));
  };

  const validateForm = (): boolean => {
    const newErrors: Record<string, string> = {};

    if (!formData.departureCity.trim()) {
      newErrors.departureCity = 'Starting city is required';
    }

    if (!formData.destinationCity.trim()) {
      newErrors.destinationCity = 'Destination city is required';
    }

    if (!formData.startDate) {
      newErrors.startDate = 'Start date is required';
    }

    if (formData.tripDuration <= 0) {
      newErrors.tripDuration = 'Trip duration must be at least 1 day';
    }

    setErrors(newErrors);
    return Object.keys(newErrors).length === 0;
  };

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (validateForm()) {
      generateTripMutation.mutate(formData);
    }
  };

  const handleNewTrip = () => {
    setShowResults(false);
    setTripResult(null);
  };

  if (showResults && tripResult) {
    return (
      <Box sx={{ maxWidth: 800, mx: 'auto', mt: 4, px: 2 }}>
        <Typography variant="h4" component="h1" align="center" gutterBottom>
          Your Trip Plan
        </Typography>
        <TripResults tripData={tripResult} />
        <Box sx={{ textAlign: 'center', mt: 2 }}>
          <Button
            variant="outlined"
            onClick={handleNewTrip}
            sx={{ mt: 2 }}
          >
            Plan Another Trip
          </Button>
        </Box>
      </Box>
    );
  }

  return (
    <LocalizationProvider dateAdapter={AdapterDateFns}>
      <Box sx={{ maxWidth: 800, mx: 'auto', mt: 4, px: 2 }}>
        <Typography variant="h4" component="h1" align="center" gutterBottom>
          Plan Your Dream Trip
        </Typography>
        
        <Card elevation={3}>
          <CardContent>
            <form onSubmit={handleSubmit}>
              <Stack spacing={3}>
                {/* Departure and Destination Cities Row */}
                <Stack direction={{ xs: 'column', md: 'row' }} spacing={3}>
                  <Autocomplete
                    freeSolo
                    options={departureSuggestions}
                    value={formData.departureCity}
                    onChange={(_event, newValue) => handleInputChange('departureCity', newValue || '')}
                    inputValue={formData.departureCity}
                    onInputChange={(_event, newInputValue) => handleInputChange('departureCity', newInputValue)}
                    renderInput={(params) => (
                      <TextField
                        {...params}
                        label="Starting City"
                        error={!!errors.departureCity}
                        helperText={errors.departureCity}
                        fullWidth
                      />
                    )}
                    sx={{ flex: 1 }}
                  />
                  
                  <Autocomplete
                    freeSolo
                    options={destinationSuggestions}
                    value={formData.destinationCity}
                    onChange={(_event, newValue) => handleInputChange('destinationCity', newValue || '')}
                    inputValue={formData.destinationCity}
                    onInputChange={(_event, newInputValue) => handleInputChange('destinationCity', newInputValue)}
                    renderInput={(params) => (
                      <TextField
                        {...params}
                        label="Destination City"
                        error={!!errors.destinationCity}
                        helperText={errors.destinationCity}
                        fullWidth
                      />
                    )}
                    sx={{ flex: 1 }}
                  />
                </Stack>
                
                {/* Middle Cities */}
                <Box>
                  <Typography variant="h6" gutterBottom>
                    Stopover Cities (up to 3)
                  </Typography>
                  
                  <Stack direction={{ xs: 'column', sm: 'row' }} spacing={1} alignItems="flex-start" mb={2}>
                    <Autocomplete
                      freeSolo
                      options={middleCitySuggestions.filter(city => 
                        !formData.middleCities.includes(city) &&
                        city !== formData.departureCity &&
                        city !== formData.destinationCity
                      )}
                      value={currentMiddleCity}
                      onChange={(_event, newValue) => setCurrentMiddleCity(newValue || '')}
                      inputValue={currentMiddleCity}
                      onInputChange={(_event, newInputValue) => setCurrentMiddleCity(newInputValue)}
                      sx={{ flex: 1 }}
                      renderInput={(params) => (
                        <TextField
                          {...params}
                          label="Add Stopover City"
                          fullWidth
                        />
                      )}
                    />
                    <Button
                      variant="outlined"
                      onClick={handleAddMiddleCity}
                      disabled={currentMiddleCity.trim() === '' || formData.middleCities.length >= 3}
                      sx={{ mt: { xs: 1, sm: 2.5 } }}
                    >
                      Add
                    </Button>
                  </Stack>
                  
                  {formData.middleCities.length > 0 && (
                    <Stack direction="row" spacing={1} flexWrap="wrap">
                      {formData.middleCities.map((city, index) => (
                        <Chip
                          key={index}
                          label={city}
                          onDelete={() => handleRemoveMiddleCity(index)}
                          color="primary"
                          variant="outlined"
                        />
                      ))}
                    </Stack>
                  )}
                  
                  <FormHelperText sx={{ ml: 0 }}>
                    {formData.middleCities.length}/3 cities added
                  </FormHelperText>
                </Box>
                
                {/* Start Date and Trip Duration Row */}
                <Stack direction={{ xs: 'column', md: 'row' }} spacing={3}>
                  <Box sx={{ flex: 1 }}>
                    <DatePicker
                      label="Start Date"
                      value={formData.startDate}
                      onChange={(newValue) => handleInputChange('startDate', newValue)}
                      minDate={new Date()}
                      slotProps={{
                        textField: {
                          error: !!errors.startDate,
                          helperText: errors.startDate,
                        }
                      }}
                    />
                  </Box>
                  
                  <FormControl fullWidth sx={{ flex: 1 }} error={!!errors.tripDuration}>
                    <InputLabel>Number of Days</InputLabel>
                    <Select
                      value={formData.tripDuration}
                      label="Number of Days"
                      onChange={(e) => handleInputChange('tripDuration', Number(e.target.value))}
                    >
                      {[...Array(30)].map((_, i) => (
                        <MenuItem key={i + 1} value={i + 1}>
                          {i + 1} day{i + 1 > 1 ? 's' : ''}
                        </MenuItem>
                      ))}
                    </Select>
                    {errors.tripDuration && (
                      <FormHelperText>{errors.tripDuration}</FormHelperText>
                    )}
                  </FormControl>
                </Stack>
                
                {/* Submit Button */}
                <Box sx={{ display: 'flex', justifyContent: 'center', mt: 2 }}>
                  <Button
                    type="submit"
                    variant="contained"
                    size="large"
                    disabled={generateTripMutation.isPending}
                    sx={{ px: 4, py: 1.5, minWidth: 200 }}
                    startIcon={generateTripMutation.isPending ? <CircularProgress size={20} /> : null}
                  >
                    {generateTripMutation.isPending ? 'Planning...' : 'Plan My Trip'}
                  </Button>
                </Box>
              </Stack>
            </form>
          </CardContent>
        </Card>
      </Box>
    </LocalizationProvider>
  );
};

export default TripPlanner;