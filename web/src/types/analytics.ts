export interface DayCount {
  date: string;
  visits: number;
  uniqueVisitors: number;
}

export interface AnalyticsOverview {
  totalVisits: number;
  uniqueVisitors: number;
  newVisitors: number;
  returningVisitors: number;
  byDay: DayCount[];
}

export interface PathCount {
  path: string;
  visits: number;
  uniqueVisitors: number;
}

export interface ReferrerCount {
  referrer: string;
  visits: number;
}

export interface LocationCount {
  country: string;
  city: string;
  visits: number;
}

export interface NamedCount {
  name: string;
  visits: number;
}

export interface DeviceBreakdown {
  browsers: NamedCount[];
  os: NamedCount[];
  devices: NamedCount[];
}

export interface RecentVisit {
  id: number;
  path: string;
  referrer: string;
  ip: string;
  country: string;
  city: string;
  browser: string;
  os: string;
  device: string;
  createdAt: string;
}

export interface DateRange {
  from: string;
  to: string;
}
