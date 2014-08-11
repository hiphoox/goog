
# Defining model 
create class Person extends V #creamos una clase vertice persona
create class Referrer extends E #creamos una clase edge referrer

# Creating Vertexs
create vertex Person set name = 'Misa'
create vertex Person set name = 'Beto'
create vertex Person set name = 'Nor'

# Creating Edges
create edge Referrer from (select from Person where name = 'Misa') to (select from Person where name = 'Beto')
create edge Referrer from (select from Person where name = 'Beto') to (select from Person where name = 'Nor')

# We traverse the graph by depth:
select in(),name from (traverse * from (select from Person where name = 'Nor') while $depth <= 2 )

#From vertice misa we get this result

{
    "result":
    [
        {

            "@type": "d",
            "@rid": "#-2:1",

            "@version": 0,
            "out":

            [
                "#11:1"
            ],
            "name": "Nor"


        },
        {
            "@type": "d",

            "@rid": "#-2:2",
            "@version": 0,

            "out":
            [
                "#11:2"

            ],
            "name": "Beto"
        },

        {
            "@type": "d",
            "@rid": "#-2:3",


            "@version": 0,
            "out":

            [
            ],
            "name": "Misa"

        }
    ]
}