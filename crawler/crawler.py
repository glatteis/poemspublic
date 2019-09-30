import mwclient
import mwparserfromhell
import mwcomposerfromhell
import json
import uuid

poems = []

num_tried = 0
num_succeeded = 0

def extract_poem(poem_page):
    parsed_text = mwparserfromhell.parse(poem_page.text())
    templates = parsed_text.filter_templates()
    text_data = [t for t in templates if "textdaten" in t.name.strip_code().lower()][0]

    author = ""
    if text_data.has("AUTOR"):
        author = text_data.get("AUTOR").value.strip_code()
    
    title = ""
    if text_data.has("TITEL"):
        title = text_data.get("TITEL").value.strip_code()
    
    year_written = ""
    if text_data.has("ENTSTEHUNGSJAHR"):
        year_written = text_data.get("ENTSTEHUNGSJAHR").value.strip_code()

    year_published = ""
    if text_data.has("ERSCHEINUNGSJAHR"):
        year_published = text_data.get("ERSCHEINUNGSJAHR").value.strip_code()
    
    origin = ""
    if text_data.has("HERKUNFT"):
        origin = text_data.get("HERKUNFT").value.strip_code()

    objs = [obj for obj in parsed_text.ifilter_tags(matches="poem")]

    if (len(objs) == 0):
        raise Exception(str(len(objs)) + " poems")

    # I found that when there are more than 1 poem tags, they have mostly the same content.
    # So it's okay that I just take one of them.
    obj = objs[0]

    poem = obj.contents
    references = []

    tags = poem.filter_tags(poem.RECURSE_OTHERS)

    reference_count = 0

    for node in tags:
        if node.tag.strip_code() == "ref":
            reference_count += 1
            poem.replace(node, "<sup>" + str(reference_count) + "</sup>")
            references.append(str(reference_count) + " " + node.contents.strip_code())
            continue
    
    templates = poem.filter_templates(poem.RECURSE_OTHERS)

    for node in templates:
        poem.remove(node)

    wikilinks = poem.filter_wikilinks(poem.RECURSE_OTHERS)

    for node in wikilinks:
        poem.remove(node)

    poem = mwcomposerfromhell.compose(poem)

    if len(poem) > 1000:
        print("too long!")
        return

    poems.append({
        "poem": poem,
        "title": title.strip(),
        "author": author.strip(),
        "year_written": year_written.strip(),
        "year_published": year_published.strip(),
        "origin": origin.strip(),
        "references": references,
    })
    
    print(poem)
    print()
    if len(references) > 0:
        print("\n".join(references))
        print()
    print(author.strip(), "-", title.strip())
    if year_written != "":
        print("entstanden " + year_written.strip())
    elif year_published != "":
        print("erschienen " + year_published.strip())
    if origin != "":
        print("aus " + origin.strip())
    print()


site = mwclient.Site(("https", "de.wikisource.org"))

page = site.pages["Liste der Gedichte"]

for poem_page in page.links():
    try:
        num_tried += 1
        extract_poem(poem_page)
        num_succeeded += 1
        print(num_tried, num_succeeded)
    except Exception as e:
        print(e)

for poem in poems:
    print(poem)
    output = json.dumps(poem)
    print(output)
    name = str(uuid.uuid4()) + ".json"
    f = open("../data/" + name, "w")
    f.write(output)
    f.close()
