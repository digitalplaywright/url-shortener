package main

//The resulting redis schema will be, where `shortURL` is the shortened url
//for the current entry:
//    - url:shortURL -> { "shortURL" -> shortURL,
//                        "longURL"  -> "url to redirect to"   }
//    - demoClick:shortURL   -> number of clicks for url in total
//    - demoCountry:shortURL -> hash of number of clicks from country
//    - demoRegion:shortURL  -> hash of number of clicks from region

func main() {

	NewShortenerApp().start()

}
