This repository contains the backend code for [status.why-ipv6.xyz](https://status.why-ipv6.xyz). The application checks multiple websites and their subdomains agains public DNS resolvers for AAAA records. Afterwards a static HTML page is generated to display the result.

## Contributing + Dev info

### Missing websites or website infos

You can report missing websites or domains by opening an issue. If you want to provide them yourself have a look at the developer documentation down below.

## Documentation for developers

The whole application config (categories, websites, subdomains, resolvers, etc) is stored in `config.yml`. The application `generator.go` parses the config file and generates the HTML page.

### Used DNS resolvers

Each key under `nameservers` represents a different DNS resolver provider. Each of them contains a list with their publicly reachable resolver IPs. Some of them (mostly the primary IPs) are commented out. There is no advantage checking against a primary and a secondary resolver. Instead every resolver provider should've one IPv4 and one IPv6.

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

An icon helps to identify a website.

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

YAML files are sometimes hard to format properly. You can use `yamlfmt` to format `config.yml` it properly before checking it in: `yamlfmt -w config.yml`

### Minifying HTML

When you generate the static page it's very huge. It consumes round about 1.3 MByte disk space and consists of many blank spaces which are good for a human developer working and debugging the generated HTML. A computer doesn't need them and only wastes ressources on it. That's why minifiers were developed. They remove unnecessary spaces, newlines etc to save up space.   
Minifying the HTML reduces the HTML to round about 546 KByte. That's a reduction of 59%. The benefits are simple: The page loads faster and less CPU is needed for rendering it in the browser. On the other side debugging a minified HTML isn't fun.

By default the HTML is not minified. If you pass `-minify` to the application the output will be minified. Use the normal version for developing and the minified version for production.

### Tags

I found out that some german internet providers websites or services are not reachable over IPv6 but they provide IPv6 to customers. To lift the blame a bit I've implemented the tag system which acts like categories. With tags you can add additional informations. Websites with tags attached will get the ✳️ emoji on the overview page. More details will be provided on the detail pane then. See the `config.yml` file to see it in action.
