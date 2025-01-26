import { NgModule } from '@angular/core';
import { RouterModule, Routes } from '@angular/router';
import { SearchByDescriptionComponent } from './components/search-by-description/search-by-description.component';
import { SearchBySymptomsComponent } from './components/search-by-symptoms/search-by-symptoms.component';

const routes: Routes = [
  { path: 'search-by-description', component: SearchByDescriptionComponent },
  { path: 'search-by-symptoms', component: SearchBySymptomsComponent },
  { path: '', redirectTo: '/search-by-description', pathMatch: 'full' }, // Default route
  { path: '**', redirectTo: '/search-by-description' } // Fallback route
];

@NgModule({
  imports: [RouterModule.forRoot(routes)],
  exports: [RouterModule]
})
export class AppRoutingModule {}
