import type { CreateRouteRequest, CreateRouteResponse } from '../gen';

// Trip segment and details interfaces for frontend
export interface TripSegment {
  id: number;
  city: string;
  coordinates: [number, number]; // [latitude, longitude]
  arrivalDate: string; // ISO date string
  departureDate: string; // ISO date string
  duration: number; // number of days
}

export interface TripDetails {
  id: number;
  route: TripSegment[];
  totalDistance: number; // in kilometers
  totalDays: number;
  origin: string;
  destination: string;
  stops: string[];
}

// API function to convert form data to the API request format
export const convertFormDataToApiRequest = (formData: {
  departureCity: string;
  middleCities: string[];
  destinationCity: string;
  startDate: Date | null;
  tripDuration: number;
}): CreateRouteRequest => {
  if (formData.startDate === null) {
    throw new Error('Start date is required');
  }

  const points = [
    formData.departureCity,
    ...formData.middleCities,
    formData.destinationCity,
  ].filter((point) => point.trim() !== '');

  return {
    points,
    startDate: formData.startDate.toISOString().split('T')[0], // Convert to YYYY-MM-DD format
    durationMinDays: formData.tripDuration,
    durationMaxDays: formData.tripDuration + 7, // Add buffer of 7 days maximum
  };
};

// Helper function to convert API response to frontend format
export const convertApiToTripDetails = (
  response: CreateRouteResponse,
  formData: {
    departureCity: string;
    middleCities: string[];
    destinationCity: string;
    startDate: Date | null;
    tripDuration: number;
  },
): TripDetails => {
  if (!formData.startDate) {
    throw new Error('Start date is required');
  }

  const route = response.points.map((point, index) => {
    // Convert timestamps to date strings
    const arrivalDate = new Date(point.startTimestamp * 1000).toISOString().split('T')[0];
    const departureDate = new Date(point.endTimestamp * 1000).toISOString().split('T')[0];

    // Calculate duration in days
    const arrival = new Date(point.startTimestamp * 1000);
    const departure = new Date(point.endTimestamp * 1000);
    const duration = Math.ceil((departure.getTime() - arrival.getTime()) / (1000 * 60 * 60 * 24));

    // Generate coordinates if not provided (fallback)
    const coordinates: [number, number] = point.coordinates
      ? [point.coordinates.latitude, point.coordinates.longitude]
      : [40 + Math.random() * 30, -10 + Math.random() * 30]; // Europe approximate coordinates

    return {
      id: index,
      city: point.name,
      coordinates,
      arrivalDate,
      departureDate,
      duration: Math.max(1, duration),
    };
  });

  // Determine origin and destination from the route
  const origin = route[0]?.city || formData.departureCity;
  const destination = route[route.length - 1]?.city || formData.destinationCity;
  const stops = route.slice(1, -1).map((s) => s.city); // All intermediate stops

  return {
    id: Date.now(),
    route,
    totalDistance: route.length * 500, // Placeholder calculation
    totalDays: formData.tripDuration,
    origin,
    destination,
    stops,
  };
};

import { getCities } from '../gen';

// API function to get city suggestions based on input
export const fetchCitySuggestions = async (input: string): Promise<string[]> => {
  try {
    // Get all cities from the API
    const response = await getCities();
    
    // Extract cities array from response (the API returns { cities: [] })
    const allCities = response.cities || [];
    
    if (!input) return allCities.slice(0, 10); // Return first 10 if no input
    
    const lowerInput = input.toLowerCase();
    return allCities.filter((city) => city?.toLowerCase().includes(lowerInput)).slice(0, 5); // Return max 5 suggestions
  } catch (error) {
    console.error('Error fetching city suggestions:', error);
    // Fallback to a small static list if API fails
    const fallbackCities = [
      'Paris',
      'London',
      'Rome',
      'Madrid',
      'Berlin',
      'Amsterdam',
      'Vienna',
      'Prague',
      'Barcelona',
      'Milan',
    ];
    
    if (!input) return fallbackCities;
    
    const lowerInput = input.toLowerCase();
    return fallbackCities.filter((city) => city.toLowerCase().includes(lowerInput)).slice(0, 5);
  }
};
