This repository contains the backend code for [ipv6-overview.xyz](https://ipv6-overview.xyz). The application checks multiple websites and their subdomains agains public DNS resolvers for AAAA records. Afterwards a static HTML page is generated to display the result.

## Contributing + Dev info

### Missing websites or website infos

You can report missing websites or domains by opening an issue. If you want to provide them yourself have a look at the developer documentation down below.

## Documentation for developers

The application consists of three different parts:

- `config.yml`: Application config like websites, resolvers, etc
- `index.html.gohtml`: Template for the HTML page
- `generator.go`: Application querying the resolvers, processing the results and rendering the HTML page

### Categories

The variable `categories` contains a list. Each item represents a category. If you want to add a new category just extend the list. Each category should have at least three websites. Never delte the categoroy `Uncategorized`. It's used as fallback when a website is not part of any category.   
There is only a loose coupling between categories and websites. So make sure that the spelling and case is correct.

### Websites

The dictionary `websites` contains all websites to check. The key of the dictionary is used as the display name on the HTML page. Each website must have the following keys:

- `href`: Single value containing a link to the websites front page.
- `hosts`: A list containing at least one domain to check for AAAA records. Only include domains where website ressources are loaded from. Skip third party CND domains.

The following keys are optional:

- `icon`: Link to a FontAwesome glyph or a picture. See section `Adding icons` for more details
- `twitter`: If the website has one or more twitter handles you can add them here. Don't forget to add quotes! Multiple twitter handles are separated by whitespaces.
- `categories`: Contains a list with category names. This is used to render a host into the respective categories. A target can be part of multiple categories. If this key is missing the target will be automatically grouped into the category `Uncategorized`
- `tags`: Tag the website. See section `Tags`.

### Adding icons

An icon helps to identify a website. But please have the copyright laws in mind.

#### By using FontAwesome

If you find an icon provided by [FontAwesome](https://fontawesome.com/icons?d=gallery&s=brands) you can add it by setting the `icon` key of a target to `fa-iconname`. The page renderer will recognize the `fa`-Prefix and will create a glyph.

#### By providing an image

If you can provide an image for a website you can add it to `dist/images/`. Afterwards set the `icon` key to the filename.   
Some websites provide their official logos and their guidelins of how to use them. You can find them by using your favorite search engine. Some search terms to find them:

- `websitename logo guidelines`
- `websitename logo press center`
- `websitename logo press media`

If you find no good source try to contact the website owner (for example on Twitter) and ask them if they have guidelines and if you are allowed to use their logo.   
Please add the image source near the `icon` key and (if available) a link to the licence, guideline and/or permission. Please don't manipulate images in any way if it's forbidden.

Before you add the file to the repository make sure you optimize the file size to save bandwidth and speed up the page.   
Here is a list of tools you can use:

- `optipng` for PNG files
- `jpegoptim` for JPG/JPEG files
- [scour](https://github.com/scour-project/scour) for SVG files

Most images can be reduced by more than 80% in file size.

### Formatting config.yml

Please use `yamlfmt` to format `config.yml` before creating a commit: `yamlfmt -w config.yml` or `make format`.

### Minifying HTML

After templating the resulting HTML page consumes roughly 1.3 MBytes disk space. It contains many white spaces and newline characters which are great for humans but useless for computers. Minifiers are stripping out the unnecessary bits speeding up the browser rendering. After minification the HTML page consumes only 546 KBytes (59% space saved). By default the HTML page is not minified. You can enable minification by passing `-minify` as application argument.

### Tags

Some providers website may not be able reachable over IPv6 but they provide it to their customers. A common example for this are ISPs or cloud providers. The tag system was implemented to counter this problem. You can attach tags in the `config.yml`. Attached tags are displayed by the 🏷️ emoji to the user. More details are displayed on the detail pane.

### Category limit

To reduce load on the resolvers while developing you can limit execution to a specific category. For example `-category-limit Shopping` would only check the websites stored in the `Shopping` category.

## Updating third party sources

### Start Bootstrap

The site uses a startbootstrap theme called sb-admin-2. The source file can be [found here](https://startbootstrap.com/themes/sb-admin-2/). Make sure you copy the minified version.
