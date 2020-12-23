# tpillz-presents
Webapp and website built for client that features an online store, media sharing, and blog posts. Project will primarily consist of Golang for backend code and JavaScript for front end. App will be deployed on Heroku. Below is a rough outline of the project's features. More info will be added and ReadMe will be updated throughout the development process. 

Core Features
1. Store
  - Beat Previews (mp3)
  - Purchase options for complete audio file download (.WAV)
    - Basic lease
    - Premium lease (exclusive rights)
      - Completely remove preview from store when purchased
    - Contracts for both leases (initiate purchase -> sign contract digitally -> complete purchase)
  - Clothing
  - Payment services (Paypal, Venmo, CashApp, Stripe)
  - Coupon codes / sales 
2. Community section
  - Newsletter format (Blog with comments)
3. Media page
  - Photos
    - Events, promo, etc…
    - Add new photos
  - Premium content (future implementation)
    - Users sign up for site to view premium content.
  - Music
    - Album art with links to Spotify, Apple Music, etc…
    - DJ mixes - steaming platform links
  - Video 
    - Music videos, blog videos, etc…
    - Embed playlists from YouTube
      - Add/remove videos through YouTube
      - ‘highlights’ playlist
4. Admin page
  - Add / remove items from the store
  - product templates (clothing, music)
  - upload files (images, audio)
  - Purchase logs 
    - item, date, price, buyer
    - Stored externally (dynamoDB)
    - Total sales 
