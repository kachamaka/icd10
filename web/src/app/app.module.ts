import { NgModule } from '@angular/core';
import { BrowserModule } from '@angular/platform-browser';
import { BrowserAnimationsModule } from '@angular/platform-browser/animations';
import { FormsModule, ReactiveFormsModule } from '@angular/forms'; // For reactive form usage
import { HttpClientModule } from '@angular/common/http'; // For HttpClient
import { MatInputModule } from '@angular/material/input';
import { MatButtonModule } from '@angular/material/button';
import { MatCardModule } from '@angular/material/card';
import { MatListModule } from '@angular/material/list';
import { MatProgressSpinnerModule } from '@angular/material/progress-spinner';
import { MatToolbarModule } from '@angular/material/toolbar';

import { AppRoutingModule } from './app-routing.module';
import { AppComponent } from './app.component';
import { SearchByDescriptionComponent } from './components/search-by-description/search-by-description.component';
import { SearchBySymptomsComponent } from './components/search-by-symptoms/search-by-symptoms.component';
import { HeaderComponent } from './components/shared/header/header.component';
import { FooterComponent } from './components/shared/footer/footer.component';
import { Icd10SearchService } from './services/icd10-search.service';

@NgModule({
  declarations: [
    AppComponent,
    SearchByDescriptionComponent,
    SearchBySymptomsComponent,
    HeaderComponent,
    FooterComponent,
  ],
  imports: [
    BrowserModule,
    BrowserAnimationsModule,
    AppRoutingModule,
    FormsModule,
    ReactiveFormsModule, // For two-way data binding
    HttpClientModule, // For making HTTP requests
    MatInputModule,
    MatToolbarModule,
    MatButtonModule,
    MatCardModule,
    MatListModule,
    MatProgressSpinnerModule,
  ],
  providers: [Icd10SearchService],
  bootstrap: [AppComponent],
})
export class AppModule {}
