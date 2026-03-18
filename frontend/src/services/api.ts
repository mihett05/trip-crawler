// Mock API service for trip planning

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

// Generate random coordinates for a city
const generateMockCoordinates = (cityName: string): [number, number] => {
  // Simple hash function to generate consistent coordinates for each city
  let hash = 0;
  for (let i = 0; i < cityName.length; i++) {
    hash = cityName.charCodeAt(i) + ((hash << 5) - hash);
  }

  // Use the hash to generate latitude and longitude
  const lat = 40 + (hash % 50);  // Range roughly between 40-90
  const lng = -5 + ((hash * 7) % 20);  // Range roughly between -5 to 15
  
  return [lat, lng];
};

// Mock API function to get city suggestions based on input
export const fetchCitySuggestions = async (input: string): Promise<string[]> => {
  // Simulate network delay
  await new Promise(resolve => setTimeout(resolve, 300));
  
  const allCities = [
    'Paris', 'London', 'Rome', 'Madrid', 'Berlin', 'Amsterdam', 'Vienna', 'Prague',
    'Barcelona', 'Milan', 'Dublin', 'Lisbon', 'Athens', 'Stockholm', 'Oslo',
    'Copenhagen', 'Helsinki', 'Warsaw', 'Budapest', 'Brussels', 'Zurich', 'Geneva',
    'Monaco', 'Munich', 'Florence', 'Venice', 'Nice', 'Edinburgh', 'Cardiff',
    'New York', 'Los Angeles', 'Chicago', 'Miami', 'Las Vegas', 'San Francisco',
    'Toronto', 'Vancouver', 'Mexico City', 'Rio de Janeiro', 'Buenos Aires',
    'São Paulo', 'Tokyo', 'Seoul', 'Beijing', 'Shanghai', 'Hong Kong', 'Singapore',
    'Bangkok', 'Kuala Lumpur', 'Jakarta', 'Delhi', 'Mumbai', 'Sydney', 'Melbourne',
    'Auckland', 'Cape Town', 'Johannesburg', 'Cairo', 'Lagos', 'Nairobi'
  ];

  if (!input) return [];
  
  const lowerInput = input.toLowerCase();
  return allCities.filter(city => 
    city.toLowerCase().includes(lowerInput)
  ).slice(0, 5); // Return max 5 suggestions
};

  // Mock API function to generate trip route
export const generateTripRoute = async (formData: { departureCity: string; middleCities: string[]; destinationCity: string; startDate: Date | null; tripDuration: number }): Promise<TripDetails> => {
  // Simulate network delay
  await new Promise(resolve => setTimeout(resolve, 1500));
  
  if (formData.startDate === null) {
    throw new Error('Start date is required');
  }
  
  // Create a route with all cities including departure, middle, and destination
  const allCities = [formData.departureCity, ...formData.middleCities, formData.destinationCity];
  
  // Generate route segments
  const routeSegments = allCities.map((city, index) => {
    const [lat, lng] = generateMockCoordinates(city);
    
    // Calculate arrival and departure dates
    const startDate = new Date(formData.startDate!);
    const arrivalDate = new Date(startDate);
    arrivalDate.setDate(arrivalDate.getDate() + Math.floor(Math.random() * index * 2)); // Stagger arrival
    
    const departureDate = new Date(arrivalDate);
    departureDate.setDate(departureDate.getDate() + 1); // Stay for at least 1 day
    
    return {
      id: index,
      city,
      coordinates: [lat, lng] as [number, number],
      arrivalDate: arrivalDate.toISOString().split('T')[0], // Format as YYYY-MM-DD
      departureDate: departureDate.toISOString().split('T')[0],
      duration: 1 + Math.floor(Math.random() * 2) // Random stay duration between 1-2 days
    };
  });
  
  // Adjust dates to fit within the trip duration
  const totalTripDays = formData.tripDuration;
  const actualTripDays = routeSegments.length * 2; // Estimate 2 days per segment
  
  if (actualTripDays > totalTripDays) {
    // Shorten the trip by reducing stays
    routeSegments.forEach(segment => {
      const arrival = new Date(segment.arrivalDate);
      const departure = new Date(arrival);
      departure.setDate(departure.getDate() + Math.min(segment.duration, Math.floor(totalTripDays / routeSegments.length)));
      
      segment.departureDate = departure.toISOString().split('T')[0];
    });
  }
  
  return {
    id: Date.now(),
    route: routeSegments,
    totalDistance: Math.floor(Math.random() * 5000) + 1000, // Random distance between 1000-6000 km
    totalDays: totalTripDays,
    origin: formData.departureCity,
    destination: formData.destinationCity,
    stops: formData.middleCities
  };
};