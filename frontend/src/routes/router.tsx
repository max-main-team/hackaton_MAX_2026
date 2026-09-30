import { Navigate, createBrowserRouter } from 'react-router-dom'
import { Guard } from './guard'
import Login from './login'
import ErrorScreen from './error'
import Onboarding from './onboarding'
import WelcomeScreen from './welcome'
import ResumeScreen from './candidate/resume'
import InvitationsScreen from './candidate/invitations'
import ReferralScreen from './candidate/referral'
import CompanyScreen from './recruiter/company'
import VacanciesScreen from './recruiter/vacancies'
import VacancyNew from './recruiter/vacancy-new'
import VacancyEdit from './recruiter/vacancy-edit'
import Feed from './recruiter/feed'
import CandidateList from './recruiter/candidate-list'
import AllCandidates from './recruiter/all-candidates'
import MapScreen from './map'

export const router = createBrowserRouter([
  {
    path: '/',
    element: <Login />,
    errorElement: <ErrorScreen />,
  },
  {
    path: '/onboarding',
    element: (
      <Guard allowNoRole>
        <Onboarding />
      </Guard>
    ),
  },
  {
    path: '/welcome',
    element: (
      <Guard>
        <WelcomeScreen />
      </Guard>
    ),
  },
  {
    path: '/resume',
    element: (
      <Guard role="candidate">
        <ResumeScreen />
      </Guard>
    ),
  },
  {
    path: '/invitations',
    element: (
      <Guard role="candidate">
        <InvitationsScreen />
      </Guard>
    ),
  },
  {
    path: '/referral',
    element: (
      <Guard>
        <ReferralScreen />
      </Guard>
    ),
  },
  {
    path: '/map',
    element: (
      <Guard>
        <MapScreen />
      </Guard>
    ),
  },
  {
    path: '/company',
    element: (
      <Guard role="recruiter">
        <CompanyScreen />
      </Guard>
    ),
  },
  {
    path: '/company/vacancies',
    element: (
      <Guard role="recruiter">
        <VacanciesScreen />
      </Guard>
    ),
  },
  {
    path: '/company/candidates',
    element: (
      <Guard role="recruiter">
        <AllCandidates />
      </Guard>
    ),
  },
  {
    path: '/company/vacancies/new',
    element: (
      <Guard role="recruiter">
        <VacancyNew />
      </Guard>
    ),
  },
  {
    path: '/company/vacancies/:id/edit',
    element: (
      <Guard role="recruiter">
        <VacancyEdit />
      </Guard>
    ),
  },
  {
    path: '/vacancy/:id/feed',
    element: (
      <Guard role="recruiter">
        <Feed />
      </Guard>
    ),
  },
  {
    path: '/vacancy/:id/list',
    element: (
      <Guard role="recruiter">
        <CandidateList />
      </Guard>
    ),
  },
  { path: '*', element: <Navigate to="/" replace /> },
])
